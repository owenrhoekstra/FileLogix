import sys
import json
import tempfile
import os

from contextlib import asynccontextmanager

import pika
from PIL import Image
import pytesseract
from fastapi import FastAPI, UploadFile, File

from rabbitmq import Consumer, RABBITMQ_URL, EXCHANGE_NAME

consumer = Consumer(queue_name="ocr.pending", routing_keys=["ocr.pending"])


from postgres import get_connection, close_all


@consumer.handler("ocr.pending")
def handle_ocr_pending(payload: dict):
    document_id = payload["document_id"]

    files = get_files(document_id)
    if not files:
        raise RuntimeError(f"no files found for document_id={document_id}")

    texts = []
    for file_id, path in files:
        set_file_status(file_id, "processing")
        try:
            text = extract_text(path)["text"]
            texts.append(text)
            set_file_status(file_id, "complete")
        except Exception:
            set_file_status(file_id, "failed")
            raise  # let the consumer's retry/dead-letter logic handle it

    combined_text = " ".join(texts)
    write_ocr_text(document_id, combined_text)
    publish_ocr_completed(document_id, combined_text)


def get_files(document_id: str) -> list[tuple[str, str]]:
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "SELECT id, path FROM files WHERE document_id = %s ORDER BY page_number",
                (document_id,),
            )
            return cur.fetchall()


def set_file_status(file_id: str, status: str):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE files SET ocr_status = %s WHERE id = %s",
                (status, file_id),
            )
        conn.commit()


def write_ocr_text(document_id: str, text: str):
    with get_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                "UPDATE documents SET ocr_text = %s WHERE id = %s",
                (text, document_id),
            )
        conn.commit()


def publish_ocr_completed(document_id: str, text: str):
    params = pika.URLParameters(RABBITMQ_URL)
    connection = pika.BlockingConnection(params)
    channel = connection.channel()

    channel.basic_publish(
        exchange=EXCHANGE_NAME,
        routing_key="ocr.completed",
        body=json.dumps({"document_id": document_id, "text": text}),
        properties=pika.BasicProperties(
            content_type="application/json",
            delivery_mode=2,
        ),
    )
    connection.close()


@asynccontextmanager
async def lifespan(app: FastAPI):
    consumer.start()
    yield
    consumer.stop()
    close_all()


app = FastAPI(lifespan=lifespan)


@app.get("/health")
def health():
    try:
        pytesseract.get_tesseract_version()
        return {"status": "ok"}
    except Exception as e:
        from fastapi import Response
        return Response(content=str(e), status_code=503)


@app.post("/ocr")
async def ocr(image: UploadFile = File()):
    data = await image.read()

    with tempfile.NamedTemporaryFile(delete=False, suffix=".webp") as tmp:
        tmp.write(data)
        tmp_path = tmp.name

    try:
        print("Py OCR Received Request")
        result = extract_text(tmp_path)
    finally:
        os.unlink(tmp_path)

    return result


def extract_text(image_path: str, confidence_threshold: int = 30) -> dict:
    img = Image.open(image_path)
    data = pytesseract.image_to_data(img, output_type=pytesseract.Output.DICT)

    words = [
        data["text"][i]
        for i in range(len(data["text"]))
        if data["text"][i].strip() and int(data["conf"][i]) > confidence_threshold
    ]

    result = " ".join(words)
    print(result + "\n")
    return {"text": result}


# command line version
if __name__ == "__main__":
    if len(sys.argv) < 2:
        print(json.dumps({"error": "Usage: python extract_text.py <image.webp>"}))
        sys.exit(1)

    try:
        result = extract_text(sys.argv[1])
        print(json.dumps(result))
    except Exception as e:
        print(json.dumps({"error": str(e)}))
        sys.exit(1)