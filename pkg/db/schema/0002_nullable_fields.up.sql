ALTER TABLE hit UPDATE "session" = 0 WHERE "session" IS NULL;
ALTER TABLE hit UPDATE referrer = '' WHERE referrer IS NULL;
ALTER TABLE hit UPDATE referrer_name = '' WHERE referrer_name IS NULL;
ALTER TABLE hit UPDATE referrer_icon = '' WHERE referrer_icon IS NULL;
ALTER TABLE hit UPDATE utm_source = '' WHERE utm_source IS NULL;
ALTER TABLE hit UPDATE utm_medium = '' WHERE utm_medium IS NULL;
ALTER TABLE hit UPDATE utm_campaign = '' WHERE utm_campaign IS NULL;
ALTER TABLE hit UPDATE utm_content = '' WHERE utm_content IS NULL;
ALTER TABLE hit UPDATE utm_term = '' WHERE utm_term IS NULL;

ALTER TABLE "hit" MODIFY COLUMN "session" DateTime('UTC') DEFAULT 0;
ALTER TABLE "hit" MODIFY COLUMN "referrer" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "referrer_name" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "referrer_icon" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_source" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_medium" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_campaign" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_content" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_term" String DEFAULT '';
