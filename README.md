<div align="center">

# 🎮 Power 4

### A modern, real-time Connect Four web game built with Go

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE.md)
[![Server](https://img.shields.io/badge/Server-Live-success?style=for-the-badge)](http://palawi.fr/power4)

[Features](#-features) • [Quick Start](#-quick-start) • [Gameplay](#-gameplay) • [Tech Stack](#-tech-stack)

<img src="docs/home-page.png" alt="Power 4 Home Screen" width="800"/>

</div>

---

## 🎯 Features

<table>
<tr>
<td width="50%">

### 🎲 Game Modes
- **Private Rooms** – Share a code with friends
- **Random Matchmaking** – Find opponents by skill rating
- **Training Mode** – Practice against the AI
- **Friend Challenges** – Direct invites to your friends list

</td>
<td width="50%">

### 👥 Social System
- **Friends List** – Add players and manage requests
- **Real-time Updates** – Auto-refreshing friend status
- **Challenge System** – Send instant game invites
- **User Profiles** – Track stats, Elo, win rate

</td>
</tr>
<tr>
<td>

### ⚡ Real-time Gameplay
- **Live Board Updates** – No JavaScript, pure HTML refresh
- **Turn Timer** – 2-minute deadline per move
- **Rematch System** – Instant rematches with alternating colors
- **Forfeit Option** – Concede gracefully anytime

</td>
<td>

### 📊 Competitive Features
- **Elo Rating System** – Skill-based matchmaking
- **Global Leaderboard** – Top players ranked
- **Match History** – Track your wins/losses
- **Fair Play** – Anti-stalling mechanics

</td>
</tr>
</table>

---

## 🚀 Quick Start

### Prerequisites
- Go 1.21 or higher
- Git

### Installation

Clone the repository

git clone https://ytrack.learn.ynov.com/git/debaptiste/power4web

cd power4

Run the server

go run main.go



Server starts at [**http://localhost:8090**](http://localhost:8090) 🎉

### First Steps
1. Create an account (username + password)
2. Try **Training Mode** to learn the game
3. Join a **Random Game** or create a **Private Room**
4. Add friends and challenge them directly!

---

## 🎮 Gameplay

<div align="center">

<img src="docs/gameplay-demo.gif" alt="Gameplay Demo" width="600"/>

### How to Win
Connect **four discs** horizontally, vertically, or diagonally before your opponent!

</div>

### Game Rules
- 7×6 grid, discs drop to the lowest available slot
- Players alternate turns
- First to align 4 discs wins
- Draw if all 42 cells fill with no winner
- Time limit: 2 minutes per turn

---

## 🛠️ Tech Stack

<div align="center">

| Layer | Technology |
|-------|-----------|
| **Backend** | Go (stdlib `net/http`) |
| **Templates** | Go `html/template` |
| **Storage** | JSON files (users, friends, sessions) |
| **Auth** | Session cookies + CSRF tokens |
| **Styling** | Custom CSS (dark theme) |
| **Real-time** | Server-sent meta refresh (no WebSockets) |

</div>

### Why Go?
- ⚡ Lightning-fast response times
- 🔒 Built-in security with standard library
- 📦 Single binary deployment
- 🧵 Concurrent game room management

---

## 📁 Project Structure

```
power4/
├── templates/ # HTML templates
│ ├── base.tmpl
│ ├── game.tmpl
│ ├── friends.tmpl
│ └── ...
├── static/
│ ├── css/ # Stylesheets
│ └── assets/ # Images, icons
├── internal/
│ ├── auth/ # Authentication & sessions
│ ├── game/ # Game logic (board, win detection)
│ └── http/ # HTTP handlers
│ ├── gamehandler.go
│ ├── friendshandler.go
│ ├── matchhandler.go
│ └── ...
├── data/ # Persistent storage (auto-created)
│ ├── users.json
│ ├── friends.json
│ └── sessions/
└── main.go # Entry point
```

---

## 🎨 Screenshots

<div align="center">

### Home Screen
<img src="docs/home.png" alt="Home" width="400"/>

### Live Game
<img src="docs/game.png" alt="Game Board" width="400"/>

### Friends System
<img src="docs/friends.png" alt="Friends" width="400"/>

### Leaderboard
<img src="docs/leaderboard.png" alt="Leaderboard" width="400"/>

</div>

---

## 🤝 Contributing

Contributions are welcome! Feel free to:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing`)
5. Open a Pull Request

---

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE.md) file for details.

---

## 🎯 Roadmap

- [ ] WebSocket support for instant updates
- [ ] Tournament mode
- [ ] Replay system

---

<div align="center">

### ⭐ Star this repo if you enjoyed playing!

Made with ❤️ and Go / GoHTML / CSS

</div>