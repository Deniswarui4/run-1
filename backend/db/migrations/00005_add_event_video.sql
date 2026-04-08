-- +goose Up
-- Add video_url column to events table for promotional videos
ALTER TABLE events ADD COLUMN video_url VARCHAR;

-- +goose Down
-- Remove video_url column
ALTER TABLE events DROP COLUMN IF EXISTS video_url;
