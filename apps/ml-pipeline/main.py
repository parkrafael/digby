from fastapi import FastAPI, UploadFile, File
from PIL import Image
from model import CLIPEmbedder
import io

app = FastAPI()
embedder = CLIPEmbedder()


@app.get("/")
def read_root():
    return {"Hello": "World"}


@app.get("/items/{item_id}")
def read_item(item_id: int, q: str | None = None):
    return {"item_id": item_id, "q": q}


@app.post("/embed")
async def embed(file: UploadFile = File(...)):
    # image in bytes
    image_bytes = await file.read()
    # bytes -> pillow image object
    image = Image.open(io.BytesIO(image_bytes))
    # create image embeddings
    features = embedder.embed_image(image)

    return {"embedding": features.tolist()}