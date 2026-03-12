CREATE TABLE IF NOT EXISTS mixes (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NULL,
    final_strength_label VARCHAR(100) NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS mix_items (
    id UUID PRIMARY KEY,
    mix_id UUID NOT NULL REFERENCES mixes(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    percent SMALLINT NOT NULL CHECK (percent > 0 AND percent <= 100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (mix_id, product_id)
);

CREATE INDEX IF NOT EXISTS idx_mixes_name ON mixes(name);
CREATE INDEX IF NOT EXISTS idx_mixes_is_active ON mixes(is_active);
CREATE INDEX IF NOT EXISTS idx_mix_items_mix_id ON mix_items(mix_id);
CREATE INDEX IF NOT EXISTS idx_mix_items_product_id ON mix_items(product_id);

CREATE OR REPLACE FUNCTION check_mix_item_product_is_hookah()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    category_code TEXT;
BEGIN
    SELECT c.code
      INTO category_code
      FROM products p
      JOIN product_categories c ON c.id = p.category_id
     WHERE p.id = NEW.product_id;

    IF category_code IS NULL OR category_code <> 'hookah_tobacco' THEN
        RAISE EXCEPTION 'mix item product must belong to hookah_tobacco category'
            USING ERRCODE = '23514';
    END IF;

    RETURN NEW;
END;
$$;

CREATE TRIGGER trg_mix_item_only_hookah
BEFORE INSERT OR UPDATE OF product_id
ON mix_items
FOR EACH ROW
EXECUTE FUNCTION check_mix_item_product_is_hookah();

CREATE OR REPLACE FUNCTION validate_mix_integrity(target_mix_id UUID)
RETURNS VOID
LANGUAGE plpgsql
AS $$
DECLARE
    mix_exists BOOLEAN;
    items_count INTEGER;
    percent_total INTEGER;
BEGIN
    SELECT EXISTS (SELECT 1 FROM mixes WHERE id = target_mix_id)
      INTO mix_exists;

    IF NOT mix_exists THEN
        RETURN;
    END IF;

    SELECT COUNT(*), COALESCE(SUM(percent), 0)
      INTO items_count, percent_total
      FROM mix_items
     WHERE mix_id = target_mix_id;

    IF items_count < 1 THEN
        RAISE EXCEPTION 'mix must contain at least 1 item'
            USING ERRCODE = '23514';
    END IF;

    IF percent_total <> 100 THEN
        RAISE EXCEPTION 'sum of mix item percent must be 100'
            USING ERRCODE = '23514';
    END IF;
END;
$$;

CREATE OR REPLACE FUNCTION trg_validate_mix_items_integrity()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
DECLARE
    affected_mix_id UUID;
BEGIN
    IF TG_OP = 'DELETE' THEN
        affected_mix_id := OLD.mix_id;
    ELSE
        affected_mix_id := NEW.mix_id;
    END IF;

    PERFORM validate_mix_integrity(affected_mix_id);

    IF TG_OP = 'DELETE' THEN
        RETURN OLD;
    END IF;
    RETURN NEW;
END;
$$;

CREATE CONSTRAINT TRIGGER trg_mix_items_integrity
AFTER INSERT OR UPDATE OR DELETE
ON mix_items
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION trg_validate_mix_items_integrity();

CREATE OR REPLACE FUNCTION trg_validate_mix_on_mixes()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    PERFORM validate_mix_integrity(NEW.id);
    RETURN NEW;
END;
$$;

CREATE CONSTRAINT TRIGGER trg_mixes_min_items_and_percent
AFTER INSERT OR UPDATE
ON mixes
DEFERRABLE INITIALLY DEFERRED
FOR EACH ROW
EXECUTE FUNCTION trg_validate_mix_on_mixes();

