UPDATE "hit" SET "path" = '/' WHERE "path" IS NULL;
ALTER TABLE "hit" ALTER COLUMN "path" SET NOT NULL;
