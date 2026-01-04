CREATE TABLE IF NOT EXISTS recipes (
    id bigserial PRIMARY KEY,
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    description text,
    instructions jsonb NOT NULL,
    notes text,
    source_url text,
    prep_time interval,
    active_time interval,
    servings integer,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    version integer NOT NULL DEFAULT 1
);

CREATE INDEX idx_recipes_user_id ON recipes(user_id);

CREATE TABLE IF NOT EXISTS ingredients (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text UNIQUE NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS equipment (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text UNIQUE NOT NULL,
    version integer NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS recipe_ingredients (
    recipe_id bigint NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    ingredient_id bigint NOT NULL REFERENCES ingredients(id),
    quantity text NOT NULL,
    unit text NOT NULL,
    optional boolean NOT NULL DEFAULT FALSE,
    PRIMARY KEY (recipe_id, ingredient_id)
);

CREATE TABLE IF NOT EXISTS recipe_equipment (
    recipe_id bigint NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    equipment_id bigint NOT NULL REFERENCES equipment(id),
    PRIMARY KEY (recipe_id, equipment_id)
);

CREATE TABLE IF NOT EXISTS recipe_instructions (
    id bigserial PRIMARY KEY,
    recipe_id bigint NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    step_number int NOT NULL,
    instruction text NOT NULL,
    notes text
);

CREATE TYPE recipe_image_type AS ENUM ('thumbnail', 'main', 'step');

CREATE TABLE recipe_images (
    id bigserial PRIMARY KEY,
    recipe_id bigint NOT NULL REFERENCES recipes(id) ON DELETE CASCADE,
    image_url TEXT NOT NULL,
    image_type recipe_image_type NOT NULL,
    caption TEXT,
    uploaded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS recipe_instruction_images (
    instruction_id bigint NOT NULL REFERENCES recipe_instructions(id) ON DELETE CASCADE,
    image_id bigint NOT NULL REFERENCES recipe_images(id) ON DELETE CASCADE,
    PRIMARY KEY (instruction_id, image_id)
);

CREATE TABLE tags (
    id bigserial PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE recipe_tags (
    recipe_id INT REFERENCES recipes(id) ON DELETE CASCADE,
    tag_id INT REFERENCES tags(id),
    PRIMARY KEY (recipe_id, tag_id)
);
