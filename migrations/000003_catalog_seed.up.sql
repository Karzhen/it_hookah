INSERT INTO product_categories (id, code, name, description, is_active)
VALUES
    (gen_random_uuid(), 'hookah_tobacco', 'Табаки', 'Табаки для кальяна', true),
    (gen_random_uuid(), 'snack', 'Закуски', 'Закуски к заказу', true),
    (gen_random_uuid(), 'cold_drink', 'Холодные напитки', 'Безалкогольные холодные напитки', true),
    (gen_random_uuid(), 'hot_drink', 'Горячие напитки', 'Чай, кофе и другие горячие напитки', true)
ON CONFLICT (code) DO NOTHING;

INSERT INTO tobacco_strengths (id, name, level, description, is_active)
VALUES
    (gen_random_uuid(), 'light', 1, 'Легкая крепость', true),
    (gen_random_uuid(), 'medium', 2, 'Средняя крепость', true),
    (gen_random_uuid(), 'strong', 3, 'Крепкая крепость', true)
ON CONFLICT (name) DO NOTHING;

INSERT INTO tobacco_flavors (id, name, description, is_active)
VALUES
    (gen_random_uuid(), 'grape', 'Виноград', true),
    (gen_random_uuid(), 'mint', 'Мята', true),
    (gen_random_uuid(), 'watermelon', 'Арбуз', true),
    (gen_random_uuid(), 'citrus', 'Цитрус', true),
    (gen_random_uuid(), 'berry', 'Ягоды', true),
    (gen_random_uuid(), 'double_apple', 'Двойное яблоко', true)
ON CONFLICT (name) DO NOTHING;
