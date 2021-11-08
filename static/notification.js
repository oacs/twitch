window.onload = async () => {
  const ws = new WebSocket("ws://localhost:7001/chat");
  ws.onmessage = function (event) {
    const messageEvent = JSON.parse(event.data);
    console.log({ messageEvent });
    if (messageEvent.subscription) {
      createNotification(messageEvent);
    }
  };
};

// create gif with sound and text using event data
const createNotification = (event) => {
  console.log(" Hola mundo");
  const notification = document.createElement("div");
  notification.classList.add("notification");
  notification.innerHTML = `
    <div class="notification-gif">
        <img id="followVid" src="siuuu.gif" width="320" height="240"/>
    </div>
    <div class="notification-text">
      <h1>${event.event["user_name"]} just ${
    event.subscription.type.split(".")[1]
  } !!!</h1>
    </div>
  `;
  const sound = new Howl({
    src: ["siuuu.mp3"],
  });
  document.body.appendChild(notification);
  sound.play();
  console.log("hola");
  setTimeout(() => {
    console.log("adios");
    document.body.removeChild(notification);
  }, 5000);
};
