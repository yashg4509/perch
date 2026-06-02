import { Pinecone } from "@pinecone-database/pinecone";

export function pineconeClient(): Pinecone {
  const apiKey = process.env.PINECONE_API_KEY;
  if (!apiKey) {
    throw new Error("PINECONE_API_KEY missing");
  }
  return new Pinecone({ apiKey });
}

export function pineconeIndex() {
  const name = process.env.PINECONE_INDEX ?? "";
  if (!name) {
    throw new Error("PINECONE_INDEX missing");
  }
  return pineconeClient().index(name);
}
