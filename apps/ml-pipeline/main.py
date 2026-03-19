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


@app.post("/embed/image")
async def embed_image(image: UploadFile = File(...)):
    # image in bytes
    image_bytes = await image.read()
    # bytes -> pillow image object
    image_pillow = Image.open(io.BytesIO(image_bytes))
    # create image embedding
    embedding = embedder.embed_image(image_pillow)[0].tolist()

    return {"embedding": embedding}


@app.post("/embed/text")
async def embed_text(query: str):
    # create text embedding
    embedding = embedder.embed_text(query)[0].tolist()

    return {"embedding": embedding}
