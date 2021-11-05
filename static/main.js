window.onload = async () => {
  const ws = new WebSocket("ws://localhost:8080/chat");
  const emoteMap = await getEmoteMap();
  const betterttvMap = await getBetterttvMap();

  // Dollar
  const dollar = document.createElement("span");
  dollar.classList.add("dollar");
  dollar.innerText = `$`;

  // Usr path
  const usrPath = document.createElement("span");
  usrPath.classList.add("usr-path");
  usrPath.innerText = "usr";

  // Slash
  const slash = document.createElement("span");
  slash.classList.add("slash");
  slash.innerText = "/";

  // carret
  const carret = document.createElement("span");
  carret.classList.add("carret");
  carret.innerText = " > ";

  const addMessage = (chat, rawMessage) => {
    const [user, message] = rawMessage.split(":");
    const messageContainer = document.createElement("p");
    messageContainer.classList.add("typewriter");

    // Username path
    const usernamePath = document.createElement("span");
    usernamePath.classList.add("username-path");
    usernamePath.innerText = user;
    // message
    const messageText = document.createElement("span");
    messageText.classList.add("message-text");
    messageText.innerHTML = messageParser(message, emoteMap, betterttvMap);

    messageContainer.appendChild(dollar.cloneNode(true));
    messageContainer.appendChild(slash.cloneNode(true));
    messageContainer.appendChild(usrPath.cloneNode(true));
    messageContainer.appendChild(slash.cloneNode(true));
    messageContainer.appendChild(usernamePath);
    messageContainer.appendChild(carret.cloneNode(true));
    messageContainer.appendChild(messageText);
    chat.appendChild(messageContainer);

    // if there are more than 10 messages, remove the first one
    if (chat.children.length > 10) {
      chat.removeChild(chat.children[0]);
    }

    chat.lastChild.scrollIntoView();

    console.log(chat.getChildNodes());
  };

  ws.onmessage = function (event) {
    const chat = document.getElementById("chat-container");
    addMessage(chat, event.data);
  };
};

const mapBetterttvRawJson = (emotes) =>
  emotes.reduce((acc, emote) => {
    acc[emote.code] = emote.id;
    return acc;
  }, {});
// Grab emotes from betterttv api
const getBetterttvMap = async () => {
  const channel = "102965110";
  const globalEmotes = await fetch(
    "https://api.betterttv.net/3/cached/emotes/global"
  ).then((response) => response.json());
  const channelEmotes = await fetch(
    `https://api.betterttv.net/3/cached/users/twitch/${channel}`
  ).then((response) => response.json());

  return {
    ...mapBetterttvRawJson(globalEmotes),
    ...mapBetterttvRawJson(channelEmotes.sharedEmotes),
  };
};

// Grab emotes.json and map into an object
const getEmoteMap = async () => {
  return fetch("/emotes.json")
    .then((response) => response.json())
    .then((emotes) =>
      emotes.data.reduce((acc, emote) => {
        acc[emote.name] = emote.images.url_1x;
        return acc;
      }, {})
    );
};

const messageParser = (message, emoteMap, betterttvMap) => {
  const words = message.split(" ");
  return words
    .map((word) => {
      if (emoteMap[word]) {
        return `<img src="${emoteMap[word]}" />`;
      } else if (betterttvMap[word]) {
        console.log(betterttvMap[word]);
        return `<img src="https://cdn.betterttv.net/emote/${betterttvMap[word]}/1x" />`;
      }
      return word;
    })
    .join(" ");
};
