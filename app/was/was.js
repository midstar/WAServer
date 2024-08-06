/**
 * @file WEB Application Server common Javascript library
 * See {@link https://github.com/midstar/waserver/} for a full description.
 * @copyright Joel Midstjärna 2024
 * @license MIT
 */

function wasInit(appname) {
  wasPageInit("was-page-name");
}


//////////////////////////////////////////////////////////////////////////////
// Generic page handling

// First function to call. Will store the display properties of each
// page (for use later) and hide all pages except "showPage".
function wasPageInit(showPage) {
  var elemPages = document.getElementsByClassName("was-page");
  for (let elemPage of elemPages) {
    elemPage.setAttribute("old-display", getComputedStyle(elemPage, null).display);
    if (elemPage.id != showPage) {
      elemPage.style.display = "none";
    }
  }
}

// Hides all pages except "page"
function wasPageShow(page) {
  elemPages = document.getElementsByClassName("was-page");
  for (let elemPage of elemPages) {
    if (elemPage.id != page) {
      elemPage.style.display = "none";
    } else {
      elemPage.style.display = elemPage.getAttribute("old-display");
    }
  }
}

/**
 * Page handling
 */
class WASPage {
  constructor(appname) {

  }
}

/**
 * User handling
 */
class WASUser {
  constructor(appname) {

  }
}
