
## Gator CLI RSS Blog aggregator

To run the gator cli you will need to install Postgres as well as Go.

gator can be installed by running
go install gator?

### Execution

to execute gator you will run it from the command line.

###### Windows

`gator.exe` *command* \*args

###### Linux

`./gator` *command* \*args

args should be seperated by a space

### Commands

The following commands are available:

###### User Actions

- register *username* registers the name provided in the database and sets them as the currently logged in user.
- users					- list all the users, the currently logged in user will be indicated with a *
- login <username>		- sets the current user for the program.

###### Feed Actions

- feeds 					- this returns a list of feeds
- following				- lists feeds followed by current user

- addfeed <name> <url>	- this will add the feed to the feeds table and assign the current user as a follower
- follow <url>			- adds the current user as a follower of the feed provided
- unfollow <url>			- removes the feed from the list of those followed by the user


- browse <limit>			- returns the posts for the user limited to the limit provided, if no limit is provided only 2 will be returned.

- agg <idle duration>		- runs the aggregator at set intervals fetching the feeds. the time is in format(#h#m#s)

- reset					- use this command to reset the DB to a clean state
