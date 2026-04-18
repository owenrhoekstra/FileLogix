from fastapi import FastAPI, File, UploadFile

app = FastAPI()

@app.post("/ocr")
async def ocr(file: UploadFile):
    # run OCR here
    return {"text": "some parsed garbage"}