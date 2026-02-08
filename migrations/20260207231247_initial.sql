-- Create "jobs" table
CREATE TABLE "jobs" (
  "id" text NOT NULL,
  "video_id" text NULL,
  "user_id" text NOT NULL,
  "status" character varying(20) NOT NULL DEFAULT 'pending',
  "local_path" text NOT NULL,
  "unique_name" text NULL,
  "title" character varying(100) NULL,
  "description" text NULL,
  "error_message" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_jobs_deleted_at" to table: "jobs"
CREATE INDEX "idx_jobs_deleted_at" ON "jobs" ("deleted_at");
-- Create index "idx_jobs_id" to table: "jobs"
CREATE UNIQUE INDEX "idx_jobs_id" ON "jobs" ("id");
-- Create "tags" table
CREATE TABLE "tags" (
  "id" text NOT NULL,
  "name" character varying(50) NOT NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_tags_id" to table: "tags"
CREATE UNIQUE INDEX "idx_tags_id" ON "tags" ("id");
-- Create index "idx_tags_name" to table: "tags"
CREATE UNIQUE INDEX "idx_tags_name" ON "tags" ("name");
-- Create "users" table
CREATE TABLE "users" (
  "id" text NOT NULL,
  "username" character varying(100) NOT NULL,
  "password" text NOT NULL,
  "email" character varying(100) NULL,
  "refresh_token" text NULL,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id")
);
-- Create index "idx_users_deleted_at" to table: "users"
CREATE INDEX "idx_users_deleted_at" ON "users" ("deleted_at");
-- Create index "idx_users_email" to table: "users"
CREATE UNIQUE INDEX "idx_users_email" ON "users" ("email");
-- Create index "idx_users_id" to table: "users"
CREATE UNIQUE INDEX "idx_users_id" ON "users" ("id");
-- Create index "idx_users_username" to table: "users"
CREATE UNIQUE INDEX "idx_users_username" ON "users" ("username");
-- Create "videos" table
CREATE TABLE "videos" (
  "id" text NOT NULL,
  "video_url" text NOT NULL,
  "title" character varying(100) NOT NULL,
  "description" text NULL,
  "user_id" text NOT NULL,
  "duration" text NULL,
  "thumbnail_url" text NULL,
  "views" bigint NULL DEFAULT 0,
  "created_at" timestamptz NULL,
  "updated_at" timestamptz NULL,
  "deleted_at" timestamptz NULL,
  PRIMARY KEY ("id"),
  CONSTRAINT "fk_users_videos" FOREIGN KEY ("user_id") REFERENCES "users" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
-- Create index "idx_videos_deleted_at" to table: "videos"
CREATE INDEX "idx_videos_deleted_at" ON "videos" ("deleted_at");
-- Create index "idx_videos_id" to table: "videos"
CREATE UNIQUE INDEX "idx_videos_id" ON "videos" ("id");
-- Create "video_tags" table
CREATE TABLE "video_tags" (
  "video_model_id" text NOT NULL,
  "tag_id" text NOT NULL,
  PRIMARY KEY ("video_model_id", "tag_id"),
  CONSTRAINT "fk_video_tags_tag" FOREIGN KEY ("tag_id") REFERENCES "tags" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT "fk_video_tags_video_model" FOREIGN KEY ("video_model_id") REFERENCES "videos" ("id") ON UPDATE NO ACTION ON DELETE NO ACTION
);
