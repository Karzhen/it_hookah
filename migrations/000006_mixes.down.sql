DROP TRIGGER IF EXISTS trg_mixes_min_items_and_percent ON mixes;
DROP FUNCTION IF EXISTS trg_validate_mix_on_mixes();

DROP TRIGGER IF EXISTS trg_mix_items_integrity ON mix_items;
DROP FUNCTION IF EXISTS trg_validate_mix_items_integrity();

DROP TRIGGER IF EXISTS trg_mix_item_only_hookah ON mix_items;
DROP FUNCTION IF EXISTS check_mix_item_product_is_hookah();

DROP FUNCTION IF EXISTS validate_mix_integrity(UUID);

DROP TABLE IF EXISTS mix_items;
DROP TABLE IF EXISTS mixes;

