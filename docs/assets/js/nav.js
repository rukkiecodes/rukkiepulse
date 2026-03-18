// Highlight active sidebar link based on scroll position
(function () {
  const links = document.querySelectorAll('.sidebar-link[href^="#"]')
  const sections = []

  links.forEach(link => {
    const id = link.getAttribute('href').slice(1)
    const el = document.getElementById(id)
    if (el) sections.push({ id, el, link })
  })

  function onScroll() {
    const scrollY = window.scrollY + 80
    let current = sections[0]

    for (const s of sections) {
      if (s.el.offsetTop <= scrollY) current = s
    }

    links.forEach(l => l.classList.remove('active'))
    if (current) current.link.classList.add('active')
  }

  window.addEventListener('scroll', onScroll, { passive: true })
  onScroll()
})()
