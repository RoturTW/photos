
function getCookie(cname) {
  let name = cname + "=";
  let decodedCookie = decodeURIComponent(document.cookie);
  let ca = decodedCookie.split(';');
  for (let i = 0; i < ca.length; i++) {
    let c = ca[i];
    while (c.charAt(0) == ' ') {
      c = c.substring(1);
    }
    if (c.indexOf(name) == 0) {
      return c.substring(name.length, c.length);
    }
  }
  return "";
}

window.onload = async function () {
  const session = getCookie("session_id");
  if (session != "") {
    const urlParams = new URLSearchParams(window.location.search);
    const next = urlParams.get('next');
    window.location.href = next || "/";
    return;
  }

  const urlParams = new URLSearchParams(window.location.search);
  const token = urlParams.get('token');
  const key = "rotur-gate";

  if (token) {
    const validator = await fetch("https://social.rotur.dev/generate_validator?key=" + encodeURIComponent(key) + "&auth=" + encodeURIComponent(token))
      .then(v => v.json())
      .then(v => v.validator);

    const auth = await fetch("/api/auth?v=" + encodeURIComponent(validator));
    if (auth.ok) {
      const next = urlParams.get('next');
      window.location.href = next || "/";
      return;
    }
  }

  const returnTo = window.location.href;
  window.location.href = "https://rotur.dev/auth?return_to=".concat(encodeURIComponent(returnTo));
};