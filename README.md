# CSH Nominate

Web application to make it easier to manage and track E-Board election nominations

## Modifying the Database

In order to make changes to the database schema, you will have to generate a new migration and then write the SQL that will make the desired modification(s)

### Generating a Migration

Run the following command to generate a new migration:

```
goose -s create my_migration_name sql
```

This command can be broken down into individual tokens:

- `goose` invokes the CLI program for the goose migration library
- `create` indicates that you want to create a new migration
- `my_migration_name` is the name of the migration file
- `sql` is the extension that will be applied to the migration file. 
  - The allowed values are `sql` and `go`, but this project uses `sql` for migration files

The resulting migration file will follow the format: `timestamp_my_migration_name.sql`. While goose also allows migration files to be identified sequentially, it is best practice to use timestamp for your migrations in order to prevent merge conflicts.

I plan to set up a CI/CD pipeline that will convert timestamp identified migration files to sequential ones for production/

## Environment Variables
This application requires a variety of environment variables in order to allow everything to work as intended.

```
# The DBMS that this application's database runs on
GOOSE_DRIVER=

# The database connection string
GOOSE_DBSTRING=

# The directory that contains all the migration files
GOOSE_MIGRATION_DIR=

# Whether the app should treat any user as if they are an Active Member
DEV_DISABLE_ACTIVE_FILTERS=

# Whether to treat any user as if they are an E-Board member
DEV_FORCE_IS_EBOARD=

# Whether to treat any user as if they were the Chairperson
DEV_FORCE_IS_CHAIR=
```

## To Do

- [ ] Allow Chair to open nominations and start accepting nominations
- [ ] Allow all active members to submit nominations
- [ ] Allow Chair to set a day and time that nominations close
- [ ] Allow Chair to select which positions are open for nominations
  - [ ] Button to select or deselect all positions
- [ ] Allow Chair/Eboard to view submitted nominations
- [ ] Only allow current students to be nominated
  - [ ] Make the nominee field a select field
  - [ ] Allow multiple people in one nomination (for dual directorships)
- [ ] Automatically notify nominees of their nomination after nominations close
  - [ ] Pings integration?
- [ ] Allow nominees to accept or decline their nomination(s)
- [ ] Allow Chair to set a deadline for accepting/declining nominations
  - [ ] Assume any nominations that didnt get a response are declined
- [ ] Allow Chair/E-Board to see who has accepted/declined what nominations
- [ ] Allow member of the year nominations
- [ ] Allow all active members to create custom nominations 