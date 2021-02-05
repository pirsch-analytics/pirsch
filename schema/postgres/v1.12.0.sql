DROP INDEX visitor_stats_path_index;
DROP INDEX language_stats_path_index;
DROP INDEX referrer_stats_path_index;
DROP INDEX os_stats_path_index;
DROP INDEX browser_stats_path_index;
DROP INDEX hit_path_index;

CREATE INDEX visitor_stats_path_index ON visitor_stats(LOWER("path"));
CREATE INDEX language_stats_path_index ON language_stats(LOWER("path"));
CREATE INDEX referrer_stats_path_index ON referrer_stats(LOWER("path"));
CREATE INDEX os_stats_path_index ON os_stats(LOWER("path"));
CREATE INDEX browser_stats_path_index ON browser_stats(LOWER("path"));
CREATE INDEX hit_path_index ON hit(LOWER("path"));
CREATE INDEX hit_referrer_index ON hit(LOWER("referrer"));
CREATE INDEX hit_time_date_index ON hit(DATE("time"));

UPDATE "language_stats" SET "language" = LOWER("language") WHERE 1 = 1;
UPDATE "country_stats" SET "country_code" = LOWER("country_code") WHERE 1 = 1;
