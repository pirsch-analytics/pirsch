CREATE TABLE "hit" (
    id bigint NOT NULL UNIQUE,
    fingerprint varchar(32) NOT NULL,
    path varchar(2000),
    url varchar(2000),
    language varchar(10),
    user_agent varchar(200),
    ref varchar(200),
    time timestamp without time zone NOT NULL
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

CREATE TABLE "visitors_per_day" (
    id bigint NOT NULL UNIQUE,
    day date NOT NULL,
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_day_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_day_id_seq OWNED BY "visitors_per_day".id;
ALTER TABLE ONLY "visitors_per_day" ALTER COLUMN id SET DEFAULT nextval('visitors_per_day_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_day" ADD CONSTRAINT visitors_per_day_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_day_day_index ON visitors_per_day(day);

CREATE TABLE "visitors_per_hour" (
    id bigint NOT NULL UNIQUE,
    day_and_hour timestamp without time zone NOT NULL,
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_hour_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_hour_id_seq OWNED BY "visitors_per_hour".id;
ALTER TABLE ONLY "visitors_per_hour" ALTER COLUMN id SET DEFAULT nextval('visitors_per_hour_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_hour" ADD CONSTRAINT visitors_per_hour_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_hour_day_and_hour_index ON visitors_per_hour(day_and_hour);

CREATE TABLE "visitors_per_language" (
    id bigint NOT NULL UNIQUE,
    day date NOT NULL,
    language varchar(10) NOT NULL,
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_language_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_language_id_seq OWNED BY "visitors_per_language".id;
ALTER TABLE ONLY "visitors_per_language" ALTER COLUMN id SET DEFAULT nextval('visitors_per_language_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_language" ADD CONSTRAINT visitors_per_language_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_language_day_index ON visitors_per_language(day);

CREATE TABLE "visitors_per_page" (
    id bigint NOT NULL UNIQUE,
    day date NOT NULL,
    path varchar(2000) NOT NULL,
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_page_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_page_id_seq OWNED BY "visitors_per_page".id;
ALTER TABLE ONLY "visitors_per_page" ALTER COLUMN id SET DEFAULT nextval('visitors_per_page_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_page" ADD CONSTRAINT visitors_per_page_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_page_day_index ON visitors_per_page(day);
