import { verifyToken } from "@clerk/backend";

export async function requireUserId(authHeader: string | undefined): Promise<string> {
  if (!authHeader?.startsWith("Bearer ")) {
    throw new Error("unauthorized");
  }
  const token = authHeader.slice("Bearer ".length);
  const secret = process.env.CLERK_SECRET_KEY;
  if (!secret) {
    throw new Error("unauthorized");
  }
  const authorizedParties = [process.env.WEB_ORIGIN].filter((v): v is string =>
    Boolean(v)
  );

  const verified = await verifyToken(token, {
    secretKey: secret,
    ...(authorizedParties.length > 0 ? { authorizedParties } : {}),
  });
  return verified.sub;
}
