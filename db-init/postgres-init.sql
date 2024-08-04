-- Create table for topics
CREATE TABLE topics (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL
);

-- Create table for subtopics
CREATE TABLE subtopics (
    id SERIAL PRIMARY KEY,
    id_topic INTEGER NOT NULL REFERENCES topics (id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    CONSTRAINT fk_id_topic FOREIGN KEY (id_topic) REFERENCES topics (id)
);

-- Create index on id_topic in subtopics table for faster lookups
CREATE INDEX idx_subtopics_id_topic ON subtopics (id_topic);

-- Create table for videos
CREATE TABLE videos (
    id SERIAL PRIMARY KEY,
    id_subtopic INTEGER NOT NULL REFERENCES subtopics (id) ON DELETE CASCADE,
    identifier CHAR(11) NOT NULL,
    name VARCHAR(255) NOT NULL,
    CONSTRAINT fk_id_subtopic FOREIGN KEY (id_subtopic) REFERENCES subtopics (id)
);

-- Create unique index on identifier in videos table
CREATE UNIQUE INDEX idx_videos_identifier ON videos (identifier);

-- Create index on id_subtopic in videos table for faster lookups
CREATE INDEX idx_videos_id_subtopic ON videos (id_subtopic);

-- Create table for tasks
CREATE TABLE tasks (
    id SERIAL PRIMARY KEY,
    id_subtopic INTEGER NOT NULL REFERENCES subtopics (id) ON DELETE CASCADE,
    order_num INTEGER NOT NULL,
    input VARCHAR(512) NOT NULL,
    output VARCHAR(512) NOT NULL,
    input_output_example TEXT,
    is_final_boss BOOLEAN NOT NULL DEFAULT FALSE,
    starter_code TEXT,
    step_1_code TEXT,
    step_2_code TEXT,
    step_3_code TEXT,
    helper_1_text VARCHAR(255),
    helper_2_text VARCHAR(255),
    helper_3_text VARCHAR(255),
    solution_code TEXT,
    CONSTRAINT fk_id_subtopic FOREIGN KEY (id_subtopic) REFERENCES subtopics (id)
);

-- Create index on id_subtopic in tasks table for faster lookups
CREATE INDEX idx_tasks_id_subtopic ON tasks (id_subtopic);

-- Create table for tests
CREATE TABLE tests (
    id SERIAL PRIMARY KEY,
    id_task INTEGER NOT NULL REFERENCES tasks (id) ON DELETE CASCADE,
    input VARCHAR(512) NOT NULL,
    expected_output VARCHAR(512) NOT NULL,
    CONSTRAINT fk_id_task FOREIGN KEY (id_task) REFERENCES tasks (id)
);

-- Create index on id_task in tests table for faster lookups
CREATE INDEX idx_tests_id_task ON tests (id_task);

-- Ensure order_num is unique within the same subtopic
ALTER TABLE tasks
ADD CONSTRAINT unique_order_num_per_subtopic UNIQUE (id_subtopic, order_num);