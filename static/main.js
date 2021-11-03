window.onload = async () => {
  const emoteMap = await getEmoteMap();
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

  const ws = new WebSocket("ws://localhost:8080/chat");

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
    messageText.innerHTML = messageParser(message, emoteMap);

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

const messageParser = (message, emoteMap) => {
  const words = message.split(" ");
  return words
    .map((word) => {
      if (emoteMap[word]) {
        return `<img src="${emoteMap[word]}" />`;
      }
      return word;
    })
    .join(" ");
};
