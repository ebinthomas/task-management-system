#!/bin/sh

# Run migrations
export PGPASSWORD=$DB_PASSWORD
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f /app/internal/database/migrations/001_create_tasks_table.sql

# Start the application
exec ./main 