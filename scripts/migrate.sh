#!/bin/bash

# Database Migration Script for Base Service
# Usage: ./scripts/migrate.sh [command] [options]

set -e  # Exit on error

# Default configuration
DB_USER="${DB_USER:-postgres}"
DB_NAME="${DB_NAME:-orders}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
MIGRATIONS_DIR="internal/database/migrations"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

# Check if psql is available
check_psql() {
    if ! command -v psql &> /dev/null; then
        print_error "psql command not found. Please install PostgreSQL client."
        exit 1
    fi
}

# Check if database exists
check_database() {
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        print_error "Database '$DB_NAME' does not exist."
        print_info "Create it with: psql -U $DB_USER -c \"CREATE DATABASE $DB_NAME;\""
        exit 1
    fi
}

# Run SQL file
run_sql() {
    local sql_file=$1
    print_info "Running: $sql_file"

    if psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$sql_file" > /dev/null 2>&1; then
        print_success "Successfully executed: $(basename "$sql_file")"
        return 0
    else
        print_error "Failed to execute: $(basename "$sql_file")"
        return 1
    fi
}

# Initialize database (create database and run initial schema)
cmd_init() {
    print_info "Initializing database: $DB_NAME"

    # Create database if it doesn't exist
    if ! psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -lqt | cut -d \| -f 1 | grep -qw "$DB_NAME"; then
        print_info "Creating database: $DB_NAME"
        psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"
        print_success "Database created"
    else
        print_warning "Database already exists"
    fi

    # Run initial schema
    print_info "Applying initial schema..."
    run_sql "$MIGRATIONS_DIR/000_initial_schema.up.sql"

    print_success "Database initialized successfully!"
}

# Apply all migrations
cmd_up() {
    check_psql
    check_database

    print_info "Applying all migrations to: $DB_NAME"

    local migrations=($(ls "$MIGRATIONS_DIR"/*_*.up.sql | sort))

    if [ ${#migrations[@]} -eq 0 ]; then
        print_warning "No migration files found in $MIGRATIONS_DIR"
        exit 0
    fi

    local count=0
    for migration in "${migrations[@]}"; do
        if run_sql "$migration"; then
            ((count++))
        fi
    done

    print_success "Applied $count migration(s)"
}

# Rollback last migration
cmd_down() {
    check_psql
    check_database

    print_warning "Rolling back last migration from: $DB_NAME"

    local migrations=($(ls "$MIGRATIONS_DIR"/*_*.down.sql | sort -r))

    if [ ${#migrations[@]} -eq 0 ]; then
        print_error "No migration files found in $MIGRATIONS_DIR"
        exit 1
    fi

    # Get the first (most recent) migration
    local last_migration="${migrations[0]}"

    print_warning "This will run: $(basename "$last_migration")"
    read -p "Are you sure? (yes/no): " -r
    echo

    if [[ $REPLY =~ ^[Yy]es$ ]]; then
        run_sql "$last_migration"
        print_success "Rollback completed"
    else
        print_info "Rollback cancelled"
    fi
}

# Reset database (drop and recreate)
cmd_reset() {
    check_psql

    print_error "⚠️  WARNING: This will DELETE ALL DATA in database: $DB_NAME"
    print_warning "This action cannot be undone!"
    echo
    read -p "Type the database name to confirm: " -r
    echo

    if [[ $REPLY != "$DB_NAME" ]]; then
        print_info "Reset cancelled"
        exit 0
    fi

    print_info "Dropping database: $DB_NAME"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "DROP DATABASE IF EXISTS $DB_NAME;"

    print_info "Recreating database: $DB_NAME"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -c "CREATE DATABASE $DB_NAME;"

    print_info "Applying migrations..."
    cmd_up

    print_success "Database reset completed"
}

# Show migration status
cmd_status() {
    check_psql
    check_database

    print_info "Database: $DB_NAME"
    print_info "Migrations directory: $MIGRATIONS_DIR"
    echo

    print_info "Available migrations:"
    local migrations=($(ls "$MIGRATIONS_DIR"/*_*.up.sql | sort))

    for migration in "${migrations[@]}"; do
        echo "  - $(basename "$migration")"
    done

    echo
    print_info "Tables in database:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "\dt" 2>/dev/null || echo "  No tables found"

    echo
    print_info "Indices on users table:"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "SELECT indexname FROM pg_indexes WHERE tablename = 'users' ORDER BY indexname;" 2>/dev/null || echo "  Table 'users' does not exist"
}

# Show help
cmd_help() {
    cat << EOF
Database Migration Tool for Base Service

Usage:
    ./scripts/migrate.sh [command] [options]

Commands:
    init        Create database and apply initial schema
    up          Apply all pending migrations
    down        Rollback last migration
    reset       Drop and recreate database (⚠️  DELETES ALL DATA)
    status      Show migration status
    help        Show this help message

Environment Variables:
    DB_USER     Database user (default: postgres)
    DB_NAME     Database name (default: orders)
    DB_HOST     Database host (default: localhost)
    DB_PORT     Database port (default: 5432)

Examples:
    # Initialize new database
    ./scripts/migrate.sh init

    # Apply all migrations
    ./scripts/migrate.sh up

    # Check migration status
    ./scripts/migrate.sh status

    # Reset database (development only!)
    ./scripts/migrate.sh reset

    # Custom database
    DB_NAME=mydb DB_USER=myuser ./scripts/migrate.sh up

EOF
}

# Main script
main() {
    local command=${1:-help}

    case "$command" in
        init)
            cmd_init
            ;;
        up)
            cmd_up
            ;;
        down)
            cmd_down
            ;;
        reset)
            cmd_reset
            ;;
        status)
            cmd_status
            ;;
        help|--help|-h)
            cmd_help
            ;;
        *)
            print_error "Unknown command: $command"
            echo
            cmd_help
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
