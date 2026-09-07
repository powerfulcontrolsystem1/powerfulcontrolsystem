(function () {
  "use strict";

  var countryLabels = {
    PA: "Panamá",
    EC: "Ecuador"
  };

  function station(id, name, tagline, countryCode, genre, streamUrl, sourceUrl) {
    return {
      id: id,
      name: name,
      tagline: tagline,
      country: countryLabels[countryCode] || countryCode,
      countryCode: countryCode,
      genre: genre,
      streamUrl: streamUrl,
      sourceUrl: sourceUrl,
      custom: false
    };
  }

  var catalog = {
    PA: [
      station("pa-radio-maria-panama", "Radio María Panamá", "Programación católica, familiar y de acompañamiento.", "PA", "Religión / Talk", "http://dreamsiteradiocp.com:8082/stream", "http://www.radiomaria.pa/"),
      station("pa-am-original-1180", "AM Original 1180 Veraguas", "Folclor, noticias y música popular desde Veraguas.", "PA", "Folclórica / Popular", "http://rosetta.shoutca.st:9264/stream", "http://www.amoriginal.net/"),
      station("pa-original-stereo-907", "Original Stereo 90.7 FM", "Cumbia, salsa, reguetón y música tropical panameña.", "PA", "Cumbia / Salsa / Urbana", "http://rosetta.shoutca.st:8931/;", "https://originalstereo.net/"),
      station("pa-estereo-azul", "Estéreo Azul 100.9 FM", "Música tropical, salsa y entretenimiento para la jornada.", "PA", "Tropical / Salsa", "https://playerservices.streamtheworld.com/api/livestream-redirect/ESTEREOAZULAAC.aac", "https://www.estereoazul.com/"),
      station("pa-fabulosa-estereo", "Fabulosa Estéreo 100.5 FM", "Latino, urbano y programación comercial de Panamá.", "PA", "Latino / Urbano", "https://www.streaming507.net:8130/stream", "https://fabulosaestereo.com/"),
      station("pa-antena-8", "Antena 8 100.1 FM", "Hits, actualidad y radio hablada para operacion diaria.", "PA", "Hits / Talk", "https://playerservices.streamtheworld.com/api/livestream-redirect/ANT8AAC.aac", "https://www.antena8.com/"),
      station("pa-rumba-fm", "Rumba FM", "Tropical y popular para ambiente comercial activo.", "PA", "Tropical / Popular", "https://laradiossl.online:10355/;stream.nsv", "https://radioparallevar.com/"),
      station("pa-flow-927", "Flow 92.7 FM", "Reguetón, urbano y música joven de Panamá.", "PA", "Reguetón / Urbano", "http://streaming507.net:9980/stream", "https://x.com/flow927fm"),
      station("pa-fm-corazon", "FM Corazón 102.5", "Baladas y música romántica para un ambiente tranquilo.", "PA", "Romántica / Baladas", "https://sp4.colombiatelecom.com.co/8006/stream", "https://fmcorazonellenguajedelamor.com/"),
      station("pa-omega-stereo", "Omega Stereo", "Clasicos, rock suave y contenido hablado.", "PA", "Classic rock / Talk", "https://www.streaming507.net:8048/stream", "https://www.omegastereo.com/")
    ],
    EC: [
      station("ec-la-tukka", "La Tukka EC", "Música ecuatoriana, tropical y popular.", "EC", "Ecuatoriana / Tropical", "http://grupomundodigital.com:8673/live", "http://latukka.com/"),
      station("ec-la-otra-913", "Radio La Otra 91.3 FM", "Popular, farandula y entretenimiento desde Quito.", "EC", "Popular / Quito", "https://laotrafm.makrodigital.com/stream/laotrafmquito", "https://web.laotrafm.com/"),
      station("ec-radio-canela-pichincha", "Radio Canela Pichincha", "Música popular, tropical y programación nacional.", "EC", "Popular / Tropical", "https://canelaradio.makrodigital.com:9280/stream", "https://web.canelaradio.com/"),
      station("ec-radio-america-1045", "Radio America 104.5 FM", "Noticias, opinion y acompanamiento radial.", "EC", "Noticias / Radio", "https://streamingecuador.com:7030/stream?1657848016283", "https://americaestereo.com/radio-america-quito.html"),
      station("ec-radio-maria", "Radio María Ecuador", "Radio católica, espiritual y familiar.", "EC", "Religión / Talk", "http://dreamsiteradiocp4.com:8010/stream", "http://www.radiomariaecuador.org/"),
      station("ec-radio-antena-3", "Radio Antena 3 91.7 FM", "Música latina, noticias y programación regional.", "EC", "Latin / News", "https://streamingecuador.net:9368/radioantena3", "https://www.radioantena3.com/"),
      station("ec-radio-zaracay", "Radio Zaracay 100.5 FM", "Música popular y entretenimiento ecuatoriano.", "EC", "Popular", "http://stream-36.zeno.fm/as3xhhc0ts8uv?_=1", "https://www.zaracayradio.com/"),
      station("ec-onda-positiva", "Radio Onda Positiva 94.1 FM", "Latina, romántica y programación positiva.", "EC", "Latina / Romántica", "https://streamingecuador.net:8011/radioondapositiva", "https://radioondapositiva.com/"),
      station("ec-jc-la-bruja", "JC La Bruja", "Pop, entretenimiento y música actual.", "EC", "Pop", "http://s7.yesstreaming.net:8040/stream?1655753305283", "http://www.jcradio.com.ec/"),
      station("ec-radio-sucumbios", "Radio Sucumbíos 105.3 FM", "Noticias, cultura y música regional.", "EC", "Noticias / Música", "http://aler.org:8000/radiosucumbios.mp3", "https://radiosucumbios.org.ec/")
    ]
  };

  function normalizeCountry(raw) {
    var value = String(raw || "").trim().toUpperCase();
    if (value === "PANAMA") return "PA";
    if (value === "ECUADOR") return "EC";
    return catalog[value] ? value : "";
  }

  function escapeStationText(raw, fallback) {
    var value = String(raw || "").trim();
    return value || fallback || "";
  }

  function normalizeCustomStation(item, index) {
    if (!item || typeof item !== "object") return null;
    var name = escapeStationText(item.name, "");
    var streamUrl = escapeStationText(item.streamUrl, "");
    if (!name || !/^https?:\/\//i.test(streamUrl)) return null;
    var id = escapeStationText(item.id, "");
    if (!id) {
      id = "custom-" + name.toLowerCase().replace(/[^a-z0-9]+/g, "-").replace(/^-+|-+$/g, "");
    }
    if (!id || id === "custom-") id = "custom-" + String(index + 1);
    var countryCode = normalizeCountry(item.countryCode);
    return {
      id: id,
      name: name,
      tagline: escapeStationText(item.tagline, "Emisora personalizada de esta empresa."),
      country: escapeStationText(item.country, countryCode ? countryLabels[countryCode] : "Personalizada"),
      countryCode: countryCode,
      genre: escapeStationText(item.genre, "Personalizada"),
      streamUrl: streamUrl,
      sourceUrl: /^https?:\/\//i.test(String(item.sourceUrl || "")) ? String(item.sourceUrl).trim() : "",
      custom: true
    };
  }

  function normalizeCustomList(items) {
    if (!Array.isArray(items)) return [];
    return items.map(normalizeCustomStation).filter(Boolean).slice(0, 40);
  }

  function stationsForCountry(countryCode, customStations) {
    var normalized = normalizeCountry(countryCode);
    var defaults = normalized ? (catalog[normalized] || []).slice(0, 10) : [];
    var custom = normalizeCustomList(customStations);
    return defaults.concat(custom);
  }

  window.__pcsRadioCountryLabels = countryLabels;
  window.__pcsRadioSupportedCountries = ["PA", "EC"];
  window.__pcsRadioCountryCatalog = catalog;
  window.__pcsRadioStations = [];
  window.PCSRadioCatalog = {
    labels: countryLabels,
    supportedCountries: ["PA", "EC"],
    normalizeCountry: normalizeCountry,
    normalizeCustomList: normalizeCustomList,
    stationsForCountry: stationsForCountry
  };
})();
