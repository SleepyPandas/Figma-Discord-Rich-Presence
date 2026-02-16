(function () {
  const themeStorageKey = "figma-rpc-theme";
  const root = document.documentElement;
  const themeToggle = document.querySelector("[data-theme-toggle]");
  const systemThemeQuery = window.matchMedia("(prefers-color-scheme: dark)");

  const getSystemTheme = () => (systemThemeQuery.matches ? "dark" : "light");

  const getStoredTheme = () => {
    try {
      const storedTheme = localStorage.getItem(themeStorageKey);
      if (storedTheme === "light" || storedTheme === "dark") return storedTheme;
    } catch (_error) {}
    return null;
  };

  const applyTheme = (theme) => {
    root.dataset.theme = theme;
    if (!themeToggle) return;

    const isDark = theme === "dark";
    themeToggle.setAttribute("aria-pressed", String(isDark));
    themeToggle.setAttribute(
      "aria-label",
      isDark ? "Switch to light mode" : "Switch to dark mode"
    );
    themeToggle.textContent = isDark ? "Light mode" : "Dark mode";
  };

  const initialTheme = getStoredTheme() || root.dataset.theme || getSystemTheme();
  applyTheme(initialTheme);

  if (themeToggle) {
    themeToggle.addEventListener("click", () => {
      const nextTheme = root.dataset.theme === "dark" ? "light" : "dark";
      applyTheme(nextTheme);
      try {
        localStorage.setItem(themeStorageKey, nextTheme);
      } catch (_error) {}
    });
  }

  if (!getStoredTheme()) {
    const syncThemeFromSystem = (event) => {
      applyTheme(event.matches ? "dark" : "light");
    };
    if (typeof systemThemeQuery.addEventListener === "function") {
      systemThemeQuery.addEventListener("change", syncThemeFromSystem);
    } else if (typeof systemThemeQuery.addListener === "function") {
      systemThemeQuery.addListener(syncThemeFromSystem);
    }
  }

  const reduceMotion = window.matchMedia("(prefers-reduced-motion: reduce)").matches;

  const revealNodes = Array.from(document.querySelectorAll("[data-reveal]"));
  if (revealNodes.length > 0) {
    if (reduceMotion || !("IntersectionObserver" in window)) {
      revealNodes.forEach((node) => node.classList.add("is-visible"));
    } else {
      const observer = new IntersectionObserver(
        (entries) => {
          entries.forEach((entry) => {
            if (!entry.isIntersecting) return;
            entry.target.classList.add("is-visible");
            observer.unobserve(entry.target);
          });
        },
        { threshold: 0.16, rootMargin: "0px 0px -40px 0px" }
      );

      revealNodes.forEach((node, index) => {
        node.style.transitionDelay = `${Math.min(index * 70, 280)}ms`;
        observer.observe(node);
      });
    }
  }

  if (reduceMotion) return;

  const parallaxNodes = Array.from(document.querySelectorAll("[data-parallax]"));
  if (parallaxNodes.length === 0) return;

  let rafId = 0;
  const parallaxStrengthX = 52;
  const parallaxStrengthY = 38;

  const updateParallax = (clientX, clientY) => {
    const normalizedX = clientX / window.innerWidth - 0.5;
    const normalizedY = clientY / window.innerHeight - 0.5;
    const easedX = Math.sign(normalizedX) * Math.pow(Math.abs(normalizedX), 0.85);
    const easedY = Math.sign(normalizedY) * Math.pow(Math.abs(normalizedY), 0.85);

    root.style.setProperty("--cursor-x", `${clientX}px`);
    root.style.setProperty("--cursor-y", `${clientY}px`);
    root.style.setProperty("--cursor-shift-x", `${easedX * 7}px`);
    root.style.setProperty("--cursor-shift-y", `${easedY * 5}px`);

    parallaxNodes.forEach((node) => {
      const depth = Number(node.dataset.depth || "0");
      node.style.setProperty("--px", `${easedX * depth * parallaxStrengthX}px`);
      node.style.setProperty("--py", `${easedY * depth * parallaxStrengthY}px`);
    });
  };

  const handlePointerMove = (event) => {
    if (rafId) cancelAnimationFrame(rafId);
    rafId = requestAnimationFrame(() => {
      updateParallax(event.clientX, event.clientY);
      rafId = 0;
    });
  };

  window.addEventListener("pointermove", handlePointerMove, { passive: true });
  window.addEventListener(
    "pointerleave",
    () => {
      parallaxNodes.forEach((node) => {
        node.style.setProperty("--px", "0px");
        node.style.setProperty("--py", "0px");
      });
      root.style.setProperty("--cursor-shift-x", "0px");
      root.style.setProperty("--cursor-shift-y", "0px");
    },
    { passive: true }
  );
})();
