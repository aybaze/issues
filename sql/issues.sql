--
-- PostgreSQL database dump
--

-- Dumped from database version 12.0 (Debian 12.0-1.pgdg100+1)
-- Dumped by pg_dump version 12.0

-- Started on 2019-10-27 21:09:47 CET

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- TOC entry 3 (class 2615 OID 2200)
-- Name: public; Type: SCHEMA; Schema: -; Owner: postgres
--

CREATE SCHEMA public;


ALTER SCHEMA public OWNER TO postgres;

--
-- TOC entry 2924 (class 0 OID 0)
-- Dependencies: 3
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: postgres
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 205 (class 1259 OID 16400)
-- Name: relationship; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.relationship (
    "issueId" integer NOT NULL,
    "otherIssueId" integer,
    type text
);


ALTER TABLE public.relationship OWNER TO postgres;

--
-- TOC entry 204 (class 1259 OID 16392)
-- Name: servicetoken; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.servicetoken (
    "userId" integer NOT NULL,
    "accessToken" text NOT NULL,
    service text NOT NULL
);


ALTER TABLE public.servicetoken OWNER TO postgres;

--
-- TOC entry 203 (class 1259 OID 16384)
-- Name: workspace; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.workspace (
    id integer NOT NULL,
    name text,
    "repositoryIds" integer[]
);


ALTER TABLE public.workspace OWNER TO postgres;

--
-- TOC entry 2792 (class 2606 OID 16407)
-- Name: relationship relationship_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.relationship
    ADD CONSTRAINT relationship_pkey PRIMARY KEY ("issueId");


--
-- TOC entry 2790 (class 2606 OID 16399)
-- Name: servicetoken token_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.servicetoken
    ADD CONSTRAINT token_pkey PRIMARY KEY ("userId");


--
-- TOC entry 2788 (class 2606 OID 16391)
-- Name: workspace workspaces_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.workspace
    ADD CONSTRAINT workspaces_pkey PRIMARY KEY (id);


-- Completed on 2019-10-27 21:09:51 CET

--
-- PostgreSQL database dump complete
--

