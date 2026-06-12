env "local" {
  src = "file://schema.sql"
  url = env("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/voxmeet?sslmode=disable")
  dev = "docker://postgres/17/dev?search_path=public"

  migration {
    dir = "file://migrations"
  }
}
