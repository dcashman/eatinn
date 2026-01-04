ALTER TABLE recipes DROP CONSTRAINT IF EXISTS recipes_prep_time_check;

ALTER TABLE recipes DROP CONSTRAINT IF EXISTS recipes_active_time_check;

ALTER TABLE recipes DROP CONSTRAINT IF EXISTS recipes_servings_check;