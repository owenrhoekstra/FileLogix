import sys
import json
from PIL import Image
import pytesseract
import tempfile, os
from fastapi import FastAPI, UploadFile, File

app = FastAPI()

@app.post("/ocr")
async def ocr(image: UploadFile = File()):
    data = await image.read()

    with tempfile.NamedTemporaryFile(delete=False, suffix=".webp") as tmp:
        tmp.write(data)
        tmp_path = tmp.name
    
    try:
        result = extract_text(tmp_path)
    finally:
        os.unlink(tmp_path)
    
    return result


# what is called by Go backend
def extract_text(image_path: str, confidence_threshold: int = 30) -> dict:
    img = Image.open(image_path)
    data = pytesseract.image_to_data(img, output_type=pytesseract.Output.DICT)
 
    words = [
        data["text"][i]
        for i in range(len(data["text"]))
        if data["text"][i].strip() and int(data["conf"][i]) > confidence_threshold
    ]
 
    return {"text": " ".join(words)}


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