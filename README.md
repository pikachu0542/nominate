# CSH Nominate

Web application to make it easier to manage and track E-Board election nominations

## Project Structure

TODO: Update this once the project structure is finalized

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

### Homepage
- [X] Display list of positions currently accepting nominations
  - [X] Show the open and close date/time for each open position
  - [X] Button on each position listing that lets user submit a nomination
  - [X] Button on each position listing to view the results for that position
    - [X] Only display results button to Chair
- [X] Display a list of nominations the user needs to respond to
  - [X] Display response deadline for each position
  - [X] Display number of days/hours remaining tom respond
  - [X] If user has no pending nominations, hide this UI element
- [X] Link bring user to page showing all their nominations (current & past)

### My Nominations Page
- [X] Display list of nominations awaiting response
  - [X] Display names of co-nominees in the candidate
  - [X] Display response deadline for each pending nomination
  - [X] Display days/hours until response deadline
  - [X] Buttons for user to accept or decline each nomination
  - [ ] Display overall status of the candidate (determined based on responses of each member of the candidate)
- [ ] Display list of all previous nominations the user received
  - [ ] Display position
  - [ ] Display names of co-nominees
  - [ ] Display user's response
  - [ ] Display overall candidate status
  - [ ] Display date/time of user's response

### Submit Nominations Form
- [X] Allow members to submit nominations for any open position
- [ ] Restrict the submission form to only current students
- [X] Only allow current students to be nominated
- [X] Nominee selector is a dropdown containing names of current students
- [X] Allow one or more users to be selected in a single nomination

### Open Nominations Form
- [X] Ability to select one or more positions
- [ ] Button to select/deselect all positions
- [X] Ability to set closing date and time
- [X] Only allow Chair to access this form
  
### Manage Positions Page
- [X] Display a list of all positions
  - [X] State whether each position is currently accepting nominations or not
- [X] Form to create a new position
- [X] Restrict position management page to only Chair

### All Nomination Periods Page
- [X] Display list of all existing nomination periods
  - [X] Display position name for each period
  - [X] Display open date/time for each period
  - [X] Display close date/time for each period
  - [X] Display current status of each period (open/closed)
- [X] Allow user to sort list of nomination periods
  - [X] Sort by status (open first)
  - [X] Sort by position (alphabetical)
  - [X] Sort by open date (either direction)
  - [X] Sort by close date (either direction)
- [X] Allow user to filter list of nomination periods
  - [X] Filter by status (open/closed)
  - [X] Filter by position
- [X] Button to clean all sorts & filters
- [X] Restrict access to this page to only E-Board members

### Period Results Page
- [ ] Display list of submitted nominations for an open period
- [ ] Display list of candidates for a closed period (after processing by worker)
  - [ ] Show overall status of each candidate
  - [ ] Show response status of every member of a candidate
- [ ] Restrict access to this page to only Chair

### Consolidate Nominations Worker
- [ ] Insert a row into the candidates table for each unique nomination submitted for this period
- [ ] Insert a row into the candidate member table for each member of each candidate
- [ ] Insert a row into the candidate decision table for each member of each candidate, with pending as the default status
- [ ] Implement proper error handling for all potential failure points
  - [ ] Ensure error messages are meaningful (whenever possible)

### Candidate Notifications Worker
- [ ] Send a Slack DM to each member of every candidate in a newly consolidated period
  - [ ] State position user was nominated for
  - [ ] State names of co-nominees (if any)
  - [ ] State deadline to respond
  - [ ] State that nominations not responded to by the deadline will be considered declined
- [ ] Send a Slack DM to each member of a candidate when the overall candidate status changes
  - [ ] State what the status change was (old and new)
  - [ ] State why the change occurred (ie: a member of the candidate declined the nomination)
- [ ] Send reminder DMs each night to anyone who still has pending nominations
  - [ ] State what position(s) they have pending nominations for
  - [ ] State the deadline to respond for each nomination
- [ ] Update the notified_at field in the candidate table with the timestamp the notification was sent
- [ ] Implement proper error handling for all potential failure points
  - [ ] Ensure error messages are meaningful (whenever possible)

### Nomination Expiry Worker
- [ ] Identify pending nominations where the deadline to respond has passed
  - [ ] Mark each pending nomination as declined
- [ ] Implement proper error handling for all potential failure points
  - [ ] Ensure error messages are meaningful (whenever possible)

### General
- [ ] Ensure all logic has proper error handling
  - [ ] All error messages should be meaningful
- [ ] Ensure all code is sufficiently documented
- [ ] Create `Dockerfile`
- [ ] Create `docker-compose.yml`
- [ ] Document process for running locally in `README.md`
- [ ] Document all required environment variables
- [ ] Ensure `.env.template` is up to date