CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,
    code VARCHAR(100) NOT NULL UNIQUE CHECK (code ~ '^[a-z][a-z0-9_]*$'),
    name VARCHAR(150) NOT NULL,
    description TEXT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS product_tags (
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (product_id, tag_id)
);

CREATE TABLE IF NOT EXISTS mix_tags (
    mix_id UUID NOT NULL REFERENCES mixes(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (mix_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_is_active ON tags(is_active);

CREATE INDEX IF NOT EXISTS idx_product_tags_tag_id ON product_tags(tag_id);
CREATE INDEX IF NOT EXISTS idx_mix_tags_tag_id ON mix_tags(tag_id);

