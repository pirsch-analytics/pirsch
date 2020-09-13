ALTER TABLE "hit" ADD COLUMN "screen_width" integer DEFAULT 0;
ALTER TABLE "hit" ADD COLUMN "screen_height" integer DEFAULT 0;

CREATE TABLE "screen_stats" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    visitors integer NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL
);

CREATE SEQUENCE screen_stats_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE screen_stats_id_seq OWNED BY "screen_stats".id;
ALTER TABLE ONLY "screen_stats" ALTER COLUMN id SET DEFAULT nextval('screen_stats_id_seq'::regclass);
ALTER TABLE ONLY "screen_stats" ADD CONSTRAINT screen_stats_pkey PRIMARY KEY (id);
CREATE INDEX screen_stats_day_index ON screen_stats(day);
