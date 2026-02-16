(function () {
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

  const updateParallax = (clientX, clientY) => {
    const normalizedX = clientX / window.innerWidth - 0.5;
    const normalizedY = clientY / window.innerHeight - 0.5;

    parallaxNodes.forEach((node) => {
      const depth = Number(node.dataset.depth || "0");
      node.style.setProperty("--px", `${normalizedX * depth * 38}px`);
      node.style.setProperty("--py", `${normalizedY * depth * 28}px`);
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
    },
    { passive: true }
  );
})();
