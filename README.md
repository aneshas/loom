# Loom

An opinionated MVC web framework for Go — fast, simple, structured.

Goon is a fast, convention-over-configuration web framework for Go — built to get you from zero to working, deployable app with minimal setup and maximum clarity. It enforces structure, favors simplicity, and helps you ship faster without sacrificing maintainability.

# create app

Always add sqlite by default with a dummy db model and controller / view with add and list actions to illustrate the usage plus migration with a seed script - can be skipped with `--no-db` (later)

add kamal

Imitate these commands:

    [
      setup: ["deps.get", "ecto.setup", "assets.setup", "assets.build"],
      "ecto.setup": ["ecto.create", "ecto.migrate", "run priv/repo/seeds.exs"],
      "ecto.reset": ["ecto.drop", "ecto.setup"],
      test: ["ecto.create --quiet", "ecto.migrate --quiet", "test"],
      "assets.setup": ["tailwind.install --if-missing", "esbuild.install --if-missing"],
      "assets.build": ["tailwind heads_up", "esbuild heads_up"],
      "assets.deploy": [
        "tailwind heads_up --minify",
        "esbuild heads_up --minify",
        "phx.digest"
      ]
    ]

## Build / Run Watch commands

Not sure in which order:

- build templ templates
- generate controller.go and any other stuff
- generate sqlboiler models
- tailwind
- package the app as a binary + assets - optional docker image? [only on build]
- run / restart echo server [only on run - not on build]

integrate kamal from get go for easy deployment

AI support eg .claude

db create which-db:
- add migration
- add seed
- add db dep
- depending on which adapter add proper sqlboiler file and config etc ...
- db migrate

db migrate:
- run migrations
- generate sqlboiler models

scaffold:
- generate controller.go and any other stuff for a particular entity / resource
- if db is disabled skip db stuff

db migrate should run migrations and generate sqlboiler models

scaffold should generate controller, view, model, store, based on the resource name and the
generated sqlboiler models (we need to run migrations first)

it should be possible to generate only store

CI:
- add github actions with dev and prod

### MVP

- [ ] new
- [ ] deps get - installs dependencies and go mod tidy (run after new)
- sqlboiler
- air
- templ
- go mod tidy
(go tool as much as possible)

- [ ] build
- [ ] run (includes watch)
- [ ] db create (default is sqlite)
- [ ] db migrate
- [ ] db seed
- [ ] db reset
- [ ] db gen-store (generate store for a model based on sqlboiler)
- [ ] db gen-migration
- [ ] error pages - panics and errors - 404 and 500 - two layouts
- [ ] request logger with different configs for different envs
- [ ] env support
- [ ] branding / css
- [ ] auto deps
- [ ] db scaffold (minimum custom components)
- [ ] think how we can add different sets of middleware for api and html

### Later
- [ ] online docs
- [ ] tests (for loom itself and app)
- [ ] bundle assets / bootstrap
- [ ] generate auth
