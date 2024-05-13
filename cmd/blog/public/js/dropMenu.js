function dropMenu() {
  var x = document.getElementById("navbar");
  if (x.className === "navbar") {
    x.className += " dropped";
  } else {
    x.className = "navbar";
  }
}
