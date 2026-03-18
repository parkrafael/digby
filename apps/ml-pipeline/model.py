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
            features = self.model.encode_image(image_input)
        # normalizes embedding results
        features = features / features.norm(dim=-1, keepdim=True)

        return features