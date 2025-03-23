CREATE TABLE IF NOT EXISTS recipes (
    id bigserial PRIMARY KEY,  
    created_at timestamp(0) with time zone NOT NULL DEFAULT NOW(),
    name text NOT NULL,
    ingredients text[] NOT NULL, -- may want to change this to foregin key to ingredients table
    required_equipment text[] NOT NULL, -- may want to change this to foregin key to equipment table
    instructions text[] NOT NULL, -- may want to change this to jsonb
    notes text,
    display_url text,
    source_url text,
    prep_time interval,
    active_time interval,
    servings integer,
    version integer NOT NULL DEFAULT 1
);