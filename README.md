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

Track this:
https://hyperflask.dev/roadmap/

- [x] new
- [x] deps get - installs dependencies and go mod tidy (add to output prompt after new)
- sqlboiler go get -tool github.com/aarondl/sqlboiler/v4@v4.19.5
- air go get -tool github.com/air-verse/air@v1.63.0
- templ go get -tool github.com/a-h/templ/cmd/templ@v0.3.943
- go mod tidy (go tool as much as possible)
- [ ] build
- [ ] run (includes watch) - default dev - add other envs (should be a freeform name only `dev` is hardcoded)
- [ ] db migrate
- [ ] fix template map sorting bug / case sensitivity
- [ ] db seed
- [ ] db reset
- [ ] db gen-store (generate store for a model based on sqlboiler)
- [x] db gen-migration
- [ ] error pages - panics and errors - 404 and 500 - two layouts
- [ ] request logger with different configs for different envs
- [ ] env support
- [ ] branding / css
- [ ] auto deps
- [ ] think how we can add different sets of middleware for api and html since it can only be used for api for example
- [ ] cors middleware if dev mode (we can check env from loom - set it somehow when running)
- [ ] request logger middleware
- [ ] seed and migrate commands - enable env parameter (always defaults to dev)
- [ ] reset should only work with dev always
- [ ] add tpl gen (runs template generation) update air
- [ ] test db with both sqlite and postgres
- [ ] db gen-db (generates boilerplate code based on sqlboiler always regenerates) update air
- [ ] db gen-store (generates model and store based on sqlboiler if model exists exists logs and skips - add --force)
- [x] actually we could take the actual type from sqlboiler and only replace sql related types eg. nullstring to *string etc ...
- maybe we try first without the store - and map the models to templ.Attributes (map[string]interface{})
- introduce the store later or maybe only mention it in the docs as a useful pattern (in .net its similar approach with entity framework playing the role of sqlboiler and then user adds the store/repository)

- [ ] finish crud example end to end - crud and see then how the puzzle fits together
- [ ] Always add database (config, boiler toml, and main + seed)
- [ ] db scaffold contact (calls db migrate db gen-boiler gen-model and db gen-store) (minimum custom components)
- [ ] and this view model can be auto generated from sqlboiler models
- [ ] controller actions map directly from view model to db (users can add layers in the middle themselves eg. store, services, usecases, etc... - we can add generators later)

- [ ] in dev mode print errors on the page - panics etc ... custom error handler - check hyperui

### Later
- [ ] online docs
- [ ] viper config
- [ ] env support when running
- [ ] tests (for loom itself and app)
- [ ] bundle assets / bootstrap
- [ ] generate auth
- [ ] i18n
