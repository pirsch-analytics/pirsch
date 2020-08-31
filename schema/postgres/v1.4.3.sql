ALTER TABLE visitors_per_language ALTER COLUMN "language" DROP NOT NULL;
ALTER TABLE visitors_per_page ALTER COLUMN "path" DROP NOT NULL;
ALTER TABLE visitors_per_referrer ALTER COLUMN "ref" DROP NOT NULL;
