window.onload = async () => {
  const ws = new WebSocket("ws://localhost:7001/chat");
  // TODO preload gifs
  const gif = new Image();
  gif.src = "siuuu.gif";
  const sound = new Howl({
    src: ["siuuu.mp3"],
  });
  sound.play();

  ws.onmessage = function (event) {
    const messageEvent = JSON.parse(event.data);
    console.log({ messageEvent });
    if (messageEvent.subscription) {
      if (messageEvent.subscription.type !== "channel.update") {
      }
      createNotification(messageEvent, sound);
    }
  };
};

const eventTextMap = {
  "channel.update": "updated",
  "channel.subscribe": "subscribed",
  "channel.follow": "followed",
};

// create gif with sound and text using event data
const createNotification = (event, sound) => {
  const notification = document.createElement("div");
  const img = document.createElement("img");
  img.src = "siuuu.gif";
  img.style.width = "320px";
  img.style.height = "240px";
  img.classList.add("notification-gif");
  notification.classList.add("notification");

  notification.appendChild(img);
  notification.innerHTML += `
    <div class="notification-text">
      <h1>${event.event["user_name"]} just ${
    eventTextMap[event.subscription.type]
  } !!!</h1>
    </div>
  `;

  document.body.appendChild(notification);

  console.log("playin sound");
  sound.play();
  sound.on("end", () => {
    document.body.removeChild(notification);
  });
};
