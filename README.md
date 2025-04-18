
## Gator CLI RSS Blog aggregator

gator is a multi user command line application used to manage and download posts from RSS feeds.

> RSS (RDF Site Summary or Really Simple Syndication) is a web feed that allows users and applications to access updates to websites in a standardized, computer-readable format. Subscribing to RSS feeds can allow a user to keep track of many different websites in a single news aggregator, which constantly monitors sites for new content, removing the need for the user to manually check them

### Installation

to install gator you will need to install the following:
- postgresql
- Go

gator can be installed by running
`go install https://github.com/kalinith/gator`

once installed you will need to create a **.gatorconfig.json** configurations file in your home folder.
/home/*username* on Linux or c:\user\\*username* on Windows.

this file will contain the database connection string as well as the current user, you only need to edit the connection string to point to the gatorDB.

	{
	  "db_url":"postgres://postgres:postgres@localhost:5432/gator?sslmode=disable
	}

##### Database setup

This application uses PostgreSQL and Goose for database migrations.

1. Create a PostgreSQL database for the application:
   ```sql
   CREATE DATABASE gator;

Install Goose migration tool:

`go install github.com/pressly/goose/v3/cmd/goose@latest`

download the schema folder here https://github.com/kalinith/gator/tree/main/sql
and save it in your project folder.

Run the database migrations:

	cd *path/to/project/schema*
	goose postgres "postgres://username:password@localhost:5432/gator" up


This will automatically apply all migration files in order. The application includes several migrations that will create all necessary tables and seed data.

### Execution

to execute gator you will run it from the command line.

###### Windows

`gator.exe` **command** *\*args*

###### Linux

`./gator` **command** *\*args*

arguments should be seperated by a space

### Commands

The following commands are available:

##### User Actions

- **register** *username* - registers the name provided in the database and sets them as the currently logged in user.
- **users** - list all the users, the currently logged in user will be indicated with a *
- **login** *username* - sets the current user for the program.

##### Feed Actions

- **feeds**	- this returns a list of feeds
- **following** - lists feeds followed by current user
- **addfeed** *name* *url* - adds the feed to the feeds table and assign the current user as a follower
- **follow** *url* - assigns the current user as a follower of the feed provided
- **unfollow** *url* - removes the feed from the list of those followed by the user

##### Post Actions

- **browse** *limit* - returns the posts for the user limited to the limit provided, if no limit is provided only 2 will be returned.
- **agg** *idle duration* - runs the aggregator at set intervals fetching the feeds. the time is in format(#h#m#s)

##### Admin Function
- **reset** - use this command to reset the DB to a clean state

##### Examples

To create a user Kalinith
`gator register Kalinith` 
this will also log you in as Kalinith
you can then add a feed for Kalinith
`gator addfeed "Lanes Blog" "https://www.wagslane.dev/index.xml"`
