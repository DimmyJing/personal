/** @type {import("live_view_hook").Hook} */
export const MainPageContainerHook = {
  mounted() {
    /** @type {HTMLDivElement} */
    const container = document.getElementById("cursor-container");
    /** @type {HTMLDivElement} */
    const template = document.getElementById("cursor-template");
    /** @type {Set<string>} */
    const cursors = new Set();
    /** @type {Set<string>} */
    const tilesHovered = new Set();

    /** @param {HTMLDivElement} tile */
    function getTileColor(tile) {
      const tileColorMap = {
        "hover:bg-sky-300": "bg-sky-300",
        "hover:bg-pink-300": "bg-pink-300",
        "hover:bg-green-300": "bg-green-300",
        "hover:bg-yellow-300": "bg-yellow-300",
        "hover:bg-red-300": "bg-red-300",
        "hover:bg-purple-300": "bg-purple-300",
        "hover:bg-blue-300": "bg-blue-300",
        "hover:bg-indigo-300": "bg-indigo-300",
        "hover:bg-violet-300": "bg-violet-300",
      }
      let result = "";
      tile.classList.forEach(cn => {
        if (tileColorMap[cn]) {
          result = tileColorMap[cn];
        }
      });
      return result;
    }

    function randomTwoColors() {
      const colors = [
        "rgb(186 230 253)",
        "rgb(251 207 232)",
        "rgb(187 247 208)",
        "rgb(254 240 138)",
        "rgb(254 202 202)",
        "rgb(233 213 255)",
        "rgb(191 219 254)",
        "rgb(199 210 254)",
        "rgb(221 214 254)",
      ];

      const firstRandom = Math.floor(Math.random() * colors.length);
      let secondRandom = Math.floor(Math.random() * (colors.length - 1));
      if (secondRandom >= firstRandom) secondRandom++;

      return [colors[firstRandom], colors[secondRandom]];
    }

    /**
     * @param {{x: number, y: number}[]} positions
     */
    function handleUpdateCursor(positions) {
      /** @type {Set<string>} */
      const newTilesHovered = new Set();
      /** @type {Map<string, HTMLDivElement>} */
      const tilesHoveredCache = new Map();
      for (const { x, y } of positions) {
        const posx = window.innerWidth * (x / 100);
        const posy = window.innerHeight * (y / 100);
        const tileEl = document.elementFromPoint(posx, posy);
        if (tileEl && tileEl.id.startsWith("tile-")) {
          newTilesHovered.add(tileEl.id);
          tilesHoveredCache.set(tileEl.id, tileEl);
        }
      }
      const tilesToRemove = tilesHovered.difference(newTilesHovered);
      for (const tile of tilesToRemove) {
        tilesHovered.delete(tile);
        const tileEl = document.getElementById(tile);
        if (tileEl) {
          const color = getTileColor(tileEl);
          tileEl.classList.remove(color);
          tileEl.classList.remove("duration-0");
          tileEl.classList.add("duration-300");
        }
      }
      const tilesToAdd = newTilesHovered.difference(tilesHovered);
      for (const tile of tilesToAdd) {
        tilesHovered.add(tile);
        const tileEl = tilesHoveredCache.get(tile);
        if (tileEl) {
          const color = getTileColor(tileEl);
          tileEl.classList.add(color);
          tileEl.classList.add("duration-0");
          tileEl.classList.remove("duration-300");
        }
      }
    }

    this.el.addEventListener("mousemove", e => {
      const mousex = (e.pageX / window.innerWidth) * 100;
      const mousey = (e.pageY / window.innerHeight) * 100;
      this.pushEvent("main-page-mousemove", { x: mousex, y: mousey });
    });

    this.handleEvent("main-page-set-users", (payload) => {
      /** @type {{x: number, y: number, socket_id: string}[]} */
      const users = payload.users;
      /** @type {string} */
      const socketID = payload.socket_id;
      const newCursors = new Set(users.map(({ socket_id }) => socket_id));
      /** @type {Set<string>} */
      const cursorsToRemove = cursors.difference(newCursors);
      for (const cursor of cursorsToRemove) {
        /** @type {HTMLDivElement} */
        const cursorEl = document.getElementById("main-page-cursor-" + cursor);
        if (cursorEl) {
          cursorEl.remove();
        }
        cursors.delete(cursor);
      }
      /** @type {{x: number, y: number}[]} */
      const tilePositions = [];
      /** @type {Set<string>} */
      const cursorsToAdd = newCursors.difference(cursors);
      for (const cursor of cursorsToAdd) {
        if (cursor === socketID) continue;
        /** @type {HTMLDivElement} */
        const cursorEl = template.cloneNode(true);
        const user = users.find(({ socket_id }) => socket_id === cursor);
        if (!user || user.x === -1 || user.y === -1) continue;
        cursorEl.id = "main-page-cursor-" + cursor;
        cursorEl.classList.remove("hidden");
        cursorEl.classList.add("flex");
        cursorEl.style.left = `${user.x}%`;
        cursorEl.style.top = `${user.y}%`;
        const stopEls = cursorEl.getElementsByTagName("stop")
        const colors = randomTwoColors();
        stopEls[0].setAttribute("stop-color", colors[0]);
        stopEls[1].setAttribute("stop-color", colors[1]);
        const gradientEl = cursorEl.getElementsByTagName("linearGradient")[0];
        gradientEl.id = "main-page-cursor-gradient-" + cursor;
        const pathEl = cursorEl.getElementsByTagName("path")[0];
        pathEl.style.fill = `url(#main-page-cursor-gradient-${cursor})`;
        container.prepend(cursorEl);
        cursors.add(cursor);
        tilePositions.push({ x: user.x, y: user.y });
      }
      /** @type {Set<string>} */
      const cursorsToModify = cursors.intersection(newCursors);
      for (const cursor of cursorsToModify) {
        if (cursor === socketID) continue;
        /** @type {HTMLDivElement} */
        const cursorEl = document.getElementById("main-page-cursor-" + cursor);
        const user = users.find(({ socket_id }) => socket_id === cursor);
        if (!user || user.x === -1 || user.y === -1) continue;
        if (cursorEl) {
          const user = users.find(({ socket_id }) => socket_id === cursor);
          if (user) {
            cursorEl.style.left = `${user.x}%`;
            cursorEl.style.top = `${user.y}%`;
          }
        }
        tilePositions.push({ x: user.x, y: user.y });
      }

      handleUpdateCursor(tilePositions);
    });
  },
};
