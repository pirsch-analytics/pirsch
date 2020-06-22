CREATE TABLE "hit" (
    id bigint NOT NULL UNIQUE,
    fingerprint varchar(32) NOT NULL,
    path varchar(2000),
    url varchar(2000),
    language varchar(10),
    user_agent varchar(200),
    ref varchar(200),
    time timestamp with time zone
);

CREATE SEQUENCE hit_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE hit_id_seq OWNED BY "hit".id;
ALTER TABLE ONLY "hit" ALTER COLUMN id SET DEFAULT nextval('hit_id_seq'::regclass);
ALTER TABLE ONLY "hit" ADD CONSTRAINT hit_pkey PRIMARY KEY (id);
CREATE INDEX hit_fingerprint_index ON hit(fingerprint);
CREATE INDEX hit_path_index ON hit(path);
CREATE INDEX hit_time_index ON hit(time);
