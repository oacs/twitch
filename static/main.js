window.onload = async () => {
  const ws = new WebSocket("ws://localhost:7001/chat");
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
  carret.innerText = " >  ";

  const addMessage = (chat, messageEvent) => {
    const { displayName, message, tags } = messageEvent;
    const { badges } = tags;
    // split badges into an array with ,
    const badgesMap = badges.split(",").reduce((acc, badge) => {
      const [badgeName, amount] = badge.split("/");
      acc[badgeName] = amount;
      return acc;
    }, {});

    const messageContainer = document.createElement("p");
    messageContainer.classList.add("typewriter");

    // Username path
    const usernamePath = document.createElement("span");
    usernamePath.classList.add(getDisplayNameClasses(badgesMap));
    usernamePath.innerText = displayName;
    // message
    const messageText = document.createElement("span");
    messageText.classList.add("message-text");
    messageText.innerHTML = messageParser(message, emoteMap, betterttvMap);

    messageContainer.appendChild(dollar.cloneNode(true));
    messageContainer.appendChild(slash.cloneNode(true));
    messageContainer.appendChild(
      mapUsrWithBadge(badgesMap, usrPath.cloneNode(true))
    );
    messageContainer.appendChild(slash.cloneNode(true));
    messageContainer.appendChild(usernamePath);
    messageContainer.appendChild(carret.cloneNode(true));
    messageContainer.appendChild(messageText);
    chat.appendChild(messageContainer);

    // if there are more than 10 messages, remove the first one
    if (chat.children.length > 10) {
      chat.removeChild(chat.children[0]);
    }

    const messageWriter = setupTypewriter(messageText);
    messageWriter.type();

    chat.lastChild.scrollIntoView();
  };

  const chat = document.getElementById("chat-container");
  ws.onmessage = function (event) {
    const messageEvent = JSON.parse(event.data);
    if (messageEvent.type === "message") {
      addMessage(chat, messageEvent);
    }
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
        return `<img style="width: 2.3rem" src="https://cdn.betterttv.net/emote/${betterttvMap[word]}/2x" />`;
      }
      return word;
    })
    .join(" ");
};

const mapUsrWithBadge = (badges, node) => {
  const { vip, moderator, broadcaster, subscriber, premium, founder } = badges;

  if (broadcaster) {
    node.textContent = "🚀";
  } else if (moderator) {
    node.textContent = "👮";
  } else if (vip) {
    node.textContent = "💎";
  } else if (founder) {
    node.textContent = "👑";
  } else if (subscriber) {
    node.textContent = "🧑‍🚀";
  } else if (premium) {
    node.textContent = "🌙";
  }
  return node;
};

const getDisplayNameClasses = (badges) => {
  const { vip, moderator, broadcaster, subscriber, premium, founder } = badges;
  if (broadcaster) {
    return "broadcaster";
  } else if (moderator) {
    return "moderator";
  } else if (vip) {
    return "vip";
  } else if (founder) {
    return "founder";
  } else if (subscriber) {
    return "subscriber";
  } else if (premium) {
    return "premium";
  }
};

function setupTypewriter(message) {
  const innerHtml = message.innerHTML;
  const typeSpeed = 500 / innerHtml.length;

  message.innerHTML = "";

  debugger;
  let cursorPosition = 0,
    tag = document.createElement("span"),
    writingTag = false,
    tagOpen = false,
    tempTypeSpeed = 0;

  const type = function () {
    if (writingTag === true) {
      tag += innerHtml[cursorPosition];
    }

    const current = innerHtml[cursorPosition];
    if (current === "<") {
      tempTypeSpeed = 0;
      if (tagOpen) {
        tagOpen = false;
        writingTag = true;
      } else {
        tag = "";
        tagOpen = true;
        writingTag = true;
        tag += current;
      }
    }
    if (!writingTag && tagOpen) {
      tag.innerHTML += current;
    }
    if (!writingTag && !tagOpen) {
      if (current === " ") {
        tempTypeSpeed = 0;
      } else {
        tempTypeSpeed = typeSpeed;
      }
      message.innerHTML += current;
    }
    if (writingTag === true && current === ">") {
      tempTypeSpeed = typeSpeed;
      writingTag = false;
      if (tagOpen) {
        var newSpan = document.createElement("span");
        message.appendChild(newSpan);
        newSpan.innerHTML = tag;
        tag = newSpan.firstChild;
        tagOpen = false;
      }
    }

    cursorPosition += 1;
    if (cursorPosition < innerHtml.length) {
      setTimeout(type, tempTypeSpeed);
    }
  };

  return {
    type: type,
  };
}
