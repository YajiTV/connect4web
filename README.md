# Power4 — Mini serveur Go (Puissance 4)

## Prérequis

* Go ≥ 1.21

## Lancer le projet

```bash
go run ./cmd/server
```

Par défaut, le serveur écoute sur `http://localhost:8080`.

## Routes

* `GET /` : affiche la grille et l’état du jeu
* `POST /play` : joue un coup (form field `column` = 0..6), puis redirige vers `/`
* `POST /reset` : réinitialise la partie

## Structure

```
cmd/
  server/
    main.go
internal/
  game/
    board.go
  http/
    handlers.go
templates/
  base.tmpl
  index.tmpl
static/
  css/
    app.css
```