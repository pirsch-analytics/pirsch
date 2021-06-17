ALTER TABLE hit UPDATE "session" = 0 where "session" is null;
ALTER TABLE hit UPDATE referrer = '' where referrer is null;
ALTER TABLE hit UPDATE referrer_name = '' where referrer_name is null;
ALTER TABLE hit UPDATE referrer_icon = '' where referrer_icon is null;
ALTER TABLE hit UPDATE utm_source = '' where utm_source is null;
ALTER TABLE hit UPDATE utm_medium = '' where utm_medium is null;
ALTER TABLE hit UPDATE utm_campaign = '' where utm_campaign is null;
ALTER TABLE hit UPDATE utm_content = '' where utm_content is null;
ALTER TABLE hit UPDATE utm_term = '' where utm_term is null;

ALTER TABLE "hit" MODIFY COLUMN "session" DateTime('UTC') DEFAULT 0;
ALTER TABLE "hit" MODIFY COLUMN "referrer" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "referrer_name" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "referrer_icon" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_source" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_medium" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_campaign" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_content" String DEFAULT '';
ALTER TABLE "hit" MODIFY COLUMN "utm_term" String DEFAULT '';
