ALTER TABLE recipes ADD CONSTRAINT recipes_prep_time_check CHECK (prep_time >= interval '0 seconds');

ALTER TABLE recipes ADD CONSTRAINT recipes_active_time_check CHECK (active_time >= interval '0 seconds');

ALTER TABLE recipes ADD CONSTRAINT recipes_servings_check CHECK (servings > 0);