(async () => {
  console.log("ghd:: connecting to event stream");
  const evtSource = new EventSource("/_ghd/hotreload");
  evtSource.addEventListener("reload", () => {
    console.log("ghd:: reloading");
    window.location.reload();
  });
})()