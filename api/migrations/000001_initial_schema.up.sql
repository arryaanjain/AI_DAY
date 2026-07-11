CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TYPE user_status AS ENUM ('active', 'blocked', 'deleted');
CREATE TYPE auth_provider AS ENUM ('microsoft', 'phone');
CREATE TYPE module_type AS ENUM ('pixel_portrait', 'comic');
CREATE TYPE generation_status AS ENUM ('created', 'queued', 'processing', 'completed', 'failed', 'cancelled');

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    phone_e164 VARCHAR(20) UNIQUE,
    email CITEXT UNIQUE,
    display_name VARCHAR(150),
    status user_status NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_identity_present CHECK (phone_e164 IS NOT NULL OR email IS NOT NULL)
);

CREATE TABLE user_identities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider auth_provider NOT NULL,
    provider_subject VARCHAR(255) NOT NULL,
    provider_tenant VARCHAR(255),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    UNIQUE (provider, provider_subject, provider_tenant)
);
