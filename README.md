# Cloudflare cache purger

This is a simple Go app that deletes the Cloudflare cache for a specific domain.
It does so by using the Cloudflare API, authenticating using a properly set
token and the ID of the zone you want to clear.

## Setup

In Cloudflare, visit the [API Tokens](https://dash.cloudflare.com/profile/api-tokens)
section and click in the "Create Token" button. There, define the following:

* *Name:* Define a name for your token. Use something you would remember, like
    `cache-purger` or similar.
* *Permissions:* In the dropdown for "Account" select "Zone". Then, the next
    dropdown right next to "Zone", select "Cache Purge", and finally, on the
    last dropdown select "Edit". This will allow the token only to Purge Caches
    on Zones. If the token happens to be leaked, the worst thing an attacker
    could do is simply delete your cache.
* *Zones:* Here you need to decide: do you want the token to be used on all
    zones across your account? If you have 7 domains, it might be a bit
    cumbersome having to authenticate against all of them with different tokens.
    If this is still what you want to do, you can mix and match this setting:
    you can include all zones, include a few zones, exclude all zones (no idea
    why would you do that lol) or simply exclude selected zones.

Once the form is completed, submit it by clicking on "Continue to Summary".
You'll see a list of your selections here, which you can use to validate if the
configuration is correct. Click on "Create Token" if you're ready, or "Cancel"
to start over.

Additionally, you'll need the ID of the Zone on which you want to clear the
cache. This is a bit easy to get, simply select your domain from the list of
domains in the homepage (you can click the Cloudflare logo to access it directly)
then on the right side of the screen you'll see an "API" section with two IDs:
the Zone ID and the Account ID. We only need the "Zone ID", so grab it and keep
it somewhere for the setup.

## Usage

Grab one of the binaries from the [releases page](https://github.com/patrickdappollonio/cloudflare-cache-purger/releases)
and put it somewhere in your `$PATH` (or the equivalent in Windows).
Additionally, you can use the Docker image released at
`patrickdappollonio/cloudflare-cache-purger`.

In both cases, configuration is provided via environment variables:

* `TOKEN` or `INPUT_TOKEN`: The token you got following the [Setup step](#setup)
* `ZONE` or `INPUT_ZONE`: The Zone ID from the [Setup step](#setup)

While `TOKEN` and `ZONE` are the recommended, the environment variables starting
with `INPUT_` are there for compatibility and ease of use with Github Actions.

For use with Docker, simply run the container, passing the environment variables
required:

```bash
docker run -e="TOKEN=my-token" -e="ZONE=my-zone" patrickdappollonio/cloudflare-cache-purger
```

## Github Actions

You can use this project with Github actions, and as an added benefit, you'll
avoid having to pay the penalty of building the container, since it's already
published. The code for the step is:

```yaml
jobs:
  clear-cache:
    steps:
      - name: Compile and release for Windows amd64
        uses: docker://patrickdappollonio/cloudflare-cache-purger:v1.0.0
        with:
          token: ${{ secrets.CLOUDFLARE_TOKEN }}
          zone: ${{ secrets.CLOUDFLARE_ZONE }}
```

Don't forge to specify both `CLOUDFLARE_TOKEN` and `CLOUDFLARE_ZONE` as secrets
in your repository.

## Debugging

The application includes a quick-and-dirty debugging mode which shouldn't be
used in public environments, since it spills out both requests and responses,
which will contain your secrets such as tokens or zones. Enable the debug mode
by setting the environment variable `DEBUG` to any value.
