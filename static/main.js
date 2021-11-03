window.onload = () => {
  console.log("Hello World!");
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
    messageText.innerText = message;

    messageContainer.appendChild(dollar.cloneNode(true));
    messageContainer.appendChild(slash.cloneNode(true));
    messageContainer.appendChild(usrPath.cloneNode(true));
    messageContainer.appendChild(slash.cloneNode(true));
    messageContainer.appendChild(usernamePath);
    messageContainer.appendChild(carret.cloneNode(true));
    messageContainer.appendChild(messageText);
    chat.appendChild(messageContainer);
  };

  ws.onmessage = function (event) {
    const chat = document.getElementById("chat-container");
    addMessage(chat, event.data);
  };
};
