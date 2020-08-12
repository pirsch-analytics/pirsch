CREATE TABLE "visitors_per_referer" (
    id bigint NOT NULL UNIQUE,
    tenant_id bigint,
    day date NOT NULL,
    ref varchar(2000) NOT NULL,
    visitors integer NOT NULL
);

CREATE SEQUENCE visitors_per_referer_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;

ALTER SEQUENCE visitors_per_referer_id_seq OWNED BY "visitors_per_referer".id;
ALTER TABLE ONLY "visitors_per_referer" ALTER COLUMN id SET DEFAULT nextval('visitors_per_referer_id_seq'::regclass);
ALTER TABLE ONLY "visitors_per_referer" ADD CONSTRAINT visitors_per_referer_pkey PRIMARY KEY (id);
CREATE INDEX visitors_per_referer_day_index ON visitors_per_referer(day);
CREATE INDEX visitors_per_referer_tenant_id_index ON visitors_per_referer(tenant_id);
