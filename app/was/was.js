/**
 * @file WEB Application Server common Javascript library
 * See {@link https://github.com/midstar/waserver/} for a full description.
 * @copyright Joel Midstjärna 2024
 * @license MIT
 */

//////////////////////////////////////////////////////////////////////////////
// Exported globals

var WAS_APP_URL = "";

//////////////////////////////////////////////////////////////////////////////
// Private globals
var _wasAppPageCallback = null;
var _wasGameObjCallback = null;


//////////////////////////////////////////////////////////////////////////////
// Initialization
//
// appPageCallback has one parameter. Game name, will be null if not needed.
//
// appType can be following:
//  * user (Only user login)
//  * 2game (2 player game)
//
// gameObjCallback has two parameters, user and opponent. Needs to return an
// object that at least has following field:
// "players" : [user, opponent]
//
// Set to null if not a game
async function wasInit(appName, appPageCallback, appType, gameObjCallback) {
  WAS_APP_URL = `${window.location.protocol}//${window.location.host}/data/${appName}`
  _wasAppPageCallback = appPageCallback;
  _wasGameObjCallback = gameObjCallback;
  _wasPageInit("was-page-name");
  _wasPageUserInit();
  if (appType == "2game") {
    _wasPageGameSelectInit();
    _wasPageGameWaitInit();
  }
}


//////////////////////////////////////////////////////////////////////////////
// Generic page handling

// First function to call. Will store the display properties of each
// page (for use later) and hide all pages except "showPage".
function _wasPageInit(showPage) {
  var elemPages = document.getElementsByClassName("was-page");
  for (let elemPage of elemPages) {
    elemPage.setAttribute("old-display", getComputedStyle(elemPage, null).display);
    if (elemPage.id != showPage) {
      elemPage.style.display = "none";
    }
  }
}

// Hides all pages except "page"
function wasPageShow(page) {
  elemPages = document.getElementsByClassName("was-page");
  for (let elemPage of elemPages) {
    if (elemPage.id != page) {
      elemPage.style.display = "none";
    } else {
      elemPage.style.display = elemPage.getAttribute("old-display");
    }
  }
}


//////////////////////////////////////////////////////////////////////////////
// URL parameter handling

function wasSetUrlParam(param, value) {
  const params = new URLSearchParams(window.location.search);
  params.set(param, value);
  history.pushState({},"","?" + params.toString());
  //window.location.search = params.toString(); 
}

function wasGetUrlParam(param) {
  const params = new URLSearchParams(window.location.search);
  return params.get(param); 
}


//////////////////////////////////////////////////////////////////////////////
// User (login) page handling

async function _wasPageUserInit() {
  document.body.innerHTML += `
  <div id="was-page-name" class="was-page">
    <div id="was-page-name-form">
      <label for="was-username">User name:</label><br>
      <input type="text" id="was-username" name="user" oninput="_wasOnNameUpdate()"><br>
      <div class="was-button" id="was-user-login" style="display:none;" onclick="_wasOnUserLogin()">Login</div>
      <div class="was-button" id="was-user-create"  style="display:none;" onclick="_wasOnUserCreate()">Create user</div>
    </div>
  </div>
  `

  username = wasGetUrlParam("user");
  if (username != null) {
    const response = await fetch(`${WAS_APP_URL}/user/${username}`);
    if (!response.ok) {
      document.getElementById("was-username").value = username;
      _wasOnNameUpdate();
      return
    }
    _wasUserLogin(username);
  }
}

async function _wasOnNameUpdate() {
  username = document.getElementById("was-username").value;
  const response = await fetch(`${WAS_APP_URL}/user/${username}`);
  if (!response.ok) {
    document.getElementById("was-user-login").style.display = "none";
    document.getElementById("was-user-create").style.display = "flex";
    return
  }
  document.getElementById("was-user-login").style.display = "flex";
  document.getElementById("was-user-create").style.display = "none";
}

async function _wasOnUserCreate() {
  username = document.getElementById("was-username").value;
  const request = new Request(`${WAS_APP_URL}/user/${username}`, {
    method: "POST",
    body: JSON.stringify({ creationDate: new Date().toLocaleString() }),
  });
  const response = await fetch(request);
  if (!response.ok) {
    console.error(response.status);
    return
  }
  _wasUserLogin(username);
}

function _wasOnUserLogin() {
  username = document.getElementById("was-username").value;
  _wasUserLogin(username);
}

function _wasUserLogin(username) {
  wasSetUrlParam("user", username);
  if (_wasGameObjCallback == null) {
    _wasAppPageCallback(null);
  } else {
    wasShowPageGameSelect()
  }
}

//////////////////////////////////////////////////////////////////////////////
// Select (or create) game page handling

async function _wasPageGameSelectInit() {
  document.body.innerHTML += `
  <div id="was-page-game-select" class="was-page">
    <div class="was-button was-game-new" onclick="_wasOnNewGame()">New game</div>
    <div id="was-game-invites">
    </div>
  </div>
  `
}

async function wasShowPageGameSelect() {
  wasPageShow("was-page-game-select");

  // Check if we have a started game
  var gameStarted = await _wasGetGame();
  if (gameStarted != null) {
    _wasAppPageCallback(gameStarted);
    return
  }

  // Check if we have a game invite
  username = wasGetUrlParam("user");
  const response = await fetch(`${WAS_APP_URL}/game-invites/${username}`);
  if (!response.ok) {
    _wasListGameInvites();
    setTimeout(wasShowPageGameSelect, 1000);
    return
  }
  // We already have a game invite
  _wasShowPageGameWait();
}

async function _wasListGameInvites() {
  elemGameInvites = document.getElementById("was-game-invites");
  const response = await fetch(`${WAS_APP_URL}/game-invites/`);
  if (!response.ok) {
    console.log(response.status);
    return
  }
  const json = await response.json();
  elemGameInvites.innerHTML = "";
  for (var name of Object.keys(json)) {
    elemInvite = document.createElement("div");
    elemInvite.classList.add("was-button");
    elemInvite.setAttribute("onclick", `_wasOnJoinGame('${name}')`)
    elemInvite.innerText=`Join ${name}`;
    elemGameInvites.appendChild(elemInvite);
  }
}

async function _wasOnNewGame() {
  username = wasGetUrlParam("user");
  const request = new Request(`${WAS_APP_URL}/game-invites/${username}`, {
    method: "POST",
    body: JSON.stringify({ creationDate: new Date().toLocaleString() }),
  });
  const response = await fetch(request);
  if (!response.ok) {
    console.error(response.status);
    return
  }
  wasShowPageGameSelect();
}

async function _wasOnJoinGame(opponent) {
  username = wasGetUrlParam("user");

  // Remove game invite
  const request = new Request(`${WAS_APP_URL}/game-invites/${opponent}`, {
    method: "DELETE",
  });
  const response = await fetch(request);
  if (!response.ok) {
    console.error(response.status);
    return
  }
  // Start game
  wasSetUrlParam("opponent", opponent);
  _wasStartGame();
}

async function _wasStartGame() {
  username = wasGetUrlParam("user");
  opponent = wasGetUrlParam("opponent");
  var game = _wasGameObjCallback(username, opponent);
  const request = new Request(`${WAS_APP_URL}/game/${opponent}`, {
    method: "POST",
    body: JSON.stringify(game),
  });
  const response = await fetch(request);
  if (!response.ok) {
    console.error(response.status);
    return
  }
  
  wasPageShow("page-game");
  updateGamePage(opponent);
}

//////////////////////////////////////////////////////////////////////////////
// Wait game (after start invite) page handling

var _wasWaitGameStartedFlag = false;

async function _wasPageGameWaitInit() {
  document.body.innerHTML += `
  <div id="was-page-game-wait" class="was-page">
    <div>
      <div>Waiting for player to join</div>
      <div class="was-button" id="cancel-invite" onclick="_wasOnCancelInvite()">Cancel</div>
    </div>
  </div>
  `
}

async function _wasShowPageGameWait() {
  wasPageShow("was-page-game-wait");
  _wasWaitGameStartedFlag = true;
  _wasWaitGameStarted();
}

async function _wasOnCancelInvite() {
  username = wasGetUrlParam("user");
  const request = new Request(`${WAS_APP_URL}/game-invites/${username}`, {
    method: "DELETE",
  });
  const response = await fetch(request);
  if (!response.ok) {
    console.error(response.status);
    return
  }
  _wasWaitGameStartedFlag = false;
  wasShowPageGameSelect();
}

// Waits for game to start as long as _wasWaitGameStartedFlag == true
async function _wasWaitGameStarted() {
  var gameStarted = await _wasGetGame();
  if (gameStarted == null) {
    if (_wasWaitGameStartedFlag) {
      setTimeout(_wasWaitGameStarted, 1000);
    }
    return
  }
  _wasWaitGameStartedFlag = false;
  _wasAppPageCallback(gameStarted);
}

// Returns game key (id) if game has started else null
async function _wasGetGame() {
  username = wasGetUrlParam("user");
  const response = await fetch(`${WAS_APP_URL}/game/`);
  if (!response.ok) {
    console.error(response.status);
    return null
  }
  const json = await response.json();
  for ([key, game] of Object.entries(json)) {
    if (game["players"].includes(username)) {
      return key
    }
  }
  return null
}
