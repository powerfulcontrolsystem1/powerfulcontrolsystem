(function () {
  var sidebar = document.querySelector(".admin-sidebar .nav");
  var links = sidebar ? Array.from(sidebar.querySelectorAll("a")) : [];
  var iframe = document.getElementById("contentFrame");

  function clearActive() {
    links.forEach(function (a) {
      a.classList.remove("active");
    });
  }

  function setActiveByHref(href) {
    clearActive();
    var found = links.find(function (a) {
      return a.getAttribute("href") === href;
    });
    if (found) found.classList.add("active");
  }

  links.forEach(function (a) {
    a.addEventListener("click", function (e) {
      var targetAttr = a.getAttribute("target");
      if (targetAttr === "_blank" || a.classList.contains("select-company")) {
        return;
      }

      e.preventDefault();
      clearActive();
      this.classList.add("active");

      var href = a.getAttribute("href");
      if (!href) return;

      if (iframe) {
        iframe.setAttribute("src", href);
      } else {
        window.location.href = href;
      }
    });
  });

  if (iframe) {
    iframe.addEventListener("load", function () {
      try {
        var src = iframe.contentWindow.location.pathname;
        setActiveByHref(src);
      } catch (e) {
        var src2 = iframe.getAttribute("src");
        setActiveByHref(src2);
      }
    });
  }
})();

(function () {
  try {
    if (localStorage.getItem("rememberAccount") === "1") {
      fetch("/me")
        .then(function (res) {
          if (!res.ok) throw new Error("unauth");
          return res.json();
        })
        .then(function (admin) {
          if (admin && admin.email) {
            try {
              localStorage.setItem("rememberedEmail", admin.email);
            } catch (e) {}
          }
        })
        .catch(function () {});
    }
  } catch (e) {}
})();
