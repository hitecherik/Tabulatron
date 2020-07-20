# Imperial Online IV automations

These are some scripts I wrote for automating the Imperial Online IV.

## Overview

The way automations for Imperial Online IV worked was:

1. Get people to register to attend a Zoom call -- this gets you their Zoom emails.
2. Map the people's Zoom names/emails to their names/emails in Tabbycat using the `resolver` -- this generates a database.
3. The `roundrunner` uses this database to generate a Zoom breakout room CSV and a list of assignments that need to be made manually in the Zoom interface for a given round.

If you want to skip steps 1 and 2 and not have people register on Zoom, and you've imported their Zoom emails as their tabbycat emails, then the `dummyresolver` can be used instead of the resolver. If you're doing this, feel free to skip all the steps in this guide to do with Zoom setup.

## Setting this up

### Prerequisites

1. You need to be able to run `make`
2. [Go 1.14](https://golang.org/dl/) to compile the scripts
3. A [tabbycat](https://github.com/TabbycatDebate/tabbycat) instance already running
4. [Zoom JWT API keys](https://marketplace.zoom.us/docs/guides/build/jwt-app) if you're using Zoom to register your participants

### Installation

If you want them installed globally on your `PATH`, run

```bash
$ git clone https://github.com/hitecherik/Imperial-Online-IV
$ cd Imperial-Online-IV
$ sudo make install
```

If you just want to compile them in the root directory of the repository, replace the last step with just `make`.

### Configuration

Copy the `example.env` file and rename the copy to `.env`. Then, change the variables in there to correspond to your configuration:

- `TABBYCAT_API_KEY`: this is the API key for a tabbycat account with all permissions
- `TABBYCAT_URL`: this is the URL that the Tabbycat instance is hosted at
- `TABBYCAT_SLUG`: this is the short name of the tournament -- this should be visible in the URL on the public-facing tabbycat interface

The following configuration options only need to be set if you're using Zoom for registering your participants:

- `ZOOM_API_KEY`: this is the JWT API key that you created in Zoom
- `ZOOM_API_SECRET`: this is the JWT API secret that you created in Zoom
- `ZOOM_MEETING_ID`: this is the Zoom call of the meeting that your registrants are registering for

## Commands available

Below is an index of all the commands available, with a brief description of what they're for.

### `resolver`

This command builds the database that maps teams and judges to their Zoom emails based on Zoom registration data.

If you're not using Zoom to register your participants, consider using `dummyresolver` (see below) instead.

```
Usage of ./resolver:
  -db string
      JSON file to store zoom email information in (default "db.json")
  -env string
      file to read environment variables from (default ".env")
  -verbose
      print additional input
```

### `dummyresolver`

This command builds the database that maps teams and judges to their Zoom emails based on Tabbycat name and email data.

This is used to test things out or to create the database file when you're not using Zoom to register participants.

The arguments passed to `dummyresolver` are the same as the ones passed to `resolver` (see above).

### `roundrunner`

This command "runs" the given round by creating a Zoom CSV and listing allocations that need to happen manually (if there are more than 200 participants). To be able to run this command, you must have created the database using the `resolver` or `dummyresolver` first.

```
Usage of ./roundrunner:
  -csv string
      CSV file to allocate breakout rooms in (default "round.csv")
  -db string
      JSON file to store zoom email information in (default "db.json")
  -env string
      file to read environment variables from (default ".env")
  -round value
      the round to run
  -verbose
      print additional input
```

The `-round` flag takes the id of the round that is in the URL in the tabbycat admin interface, and _not_ tabbycat's internal round ID. Unless you've set tabbycat up weirdly, `-round 1` should correspond to round 1, `-round 2` to round 2, and so on.

If you're generating for multiple rounds (hint: concurrent outrounds) then the round flag can be passed multiple times (e.g.: `-round 7 -round 8`).

### `zoomregistrants`

This is a utility script that lists the people registered to the Zoom call. I created it mostly to debug issues I had with the Zoom API, although I guess it could be useful in other situations.

```
Usage of ./zoomregistrants:
  -db string
      JSON file to store registrant information in (default "registrants.json")
  -env string
      file to read environment variables from (default ".env")
  -verbose
      print additional input
```

### `tabbycatrounds`

This is another utility script that prints out each round's name and the internal tabbycat ID for that round. This was for before I realised that the internal ID doesn't correspond to the ID that needs to be passed to `roundrunner`, so this script is mostly useless.

```
Usage of ./tabbycatrounds:
  -env string
      file to read environment variables from (default ".env")
  -verbose
      print additional input
```

## Limitations

This was mostly built in one weekend, so the feature set is quite limited. In future I want to:

- Reduce some of the duplication in the code by cleaning up the structure.
- Add various integrations with Discord, the chief of which should be automatic registration.
- Fix `tabbycatrounds` so that it displays the round IDs that `roundrunner` takes as arguments.
- Add support for multiple CSVs for multiple Zoom calls.
- Find a cooler name for this project.

If you feel like helping out with any of these, fork this repo and open a PR!

## License

This project is licensed under the [MIT License](LICENSE.txt).

Copyright &copy; Alexander Nielsen, 2020.
