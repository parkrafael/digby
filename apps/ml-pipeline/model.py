import torch
import clip
from PIL import Image

class CLIPEmbedder:
    def __init__(self):
        self.device = "cpu"
        # load clip models
        self.model, self.preprocess = clip.load("ViT-B/32", device=self.device)

    def embed_image(self, image: Image.Image):
        # resizes & normalizes image -> adds batch dimension -> adds to cpu
        image_input = self.preprocess(image).unsqueeze(0).to(self.device)

        # skips gradient tracking
        with torch.no_grad():
            embedding = self.model.encode_image(image_input)
        # normalizes embedding results
        embedding = embedding / embedding.norm(dim=-1, keepdim=True)

        return embedding

    def embed_text(self, query: str):
        # tokenizes query
        query_input = clip.tokenize([query])

        # skips gradient tracking
        with torch.no_grad():
            embedding = self.model.encode_text(query_input)
        # normalizes embedding results
        embedding = embedding / embedding.norm(dim=-1, keepdim=True)

        return embedding
