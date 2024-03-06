defmodule AppWeb.PageLive do
  alias AppWeb.Presence
  use AppWeb, :live_view

  @channel_topic "cursor_page"

  @colors [
    "hover:bg-sky-300",
    "hover:bg-pink-300",
    "hover:bg-green-300",
    "hover:bg-yellow-300",
    "hover:bg-red-300",
    "hover:bg-purple-300",
    "hover:bg-blue-300",
    "hover:bg-indigo-300",
    "hover:bg-violet-300"
  ]
  defp get_colors() do
    @colors
  end

  def render(assigns) do
    ~H"""
    <.flash_group flash={@flash} />
    <ul class="list-none" id="cursor-container">
      <li
        class="pointer-events-none absolute z-30 hidden -translate-x-1 -translate-y-1 flex-col overflow-hidden whitespace-nowrap"
        id="cursor-template"
      >
        <svg xmlns="http://www.w3.org/2000/svg" width="31" height="32" fill="none" viewBox="0 0 31 32">
          <path
            fill="url(#a)"
            d="m.609 10.86 5.234 15.488c1.793 5.306 8.344 7.175 12.666 3.612l9.497-7.826c4.424-3.646 3.69-10.625-1.396-13.27L11.88 1.2C5.488-2.124-1.697 4.033.609 10.859Z"
          />
          <defs>
            <linearGradient
              id="a"
              x1="-4.982"
              x2="23.447"
              y1="-8.607"
              y2="25.891"
              gradientUnits="userSpaceOnUse"
            >
              <stop />
              <stop offset="1" />
            </linearGradient>
          </defs>
        </svg>
      </li>
    </ul>
    <div
      id="main-page-container"
      class="relative flex h-[100vh] flex-col items-center justify-center overflow-hidden bg-slate-900"
      phx-hook="MainPageContainer"
    >
      <div class="pointer-events-none absolute inset-0 z-20 h-full w-full bg-slate-900 [mask-image:radial-gradient(transparent,white)]">
      </div>

      <div class="absolute -top-1/4 left-1/4 z-0 flex h-full w-full -translate-x-[40%] -translate-y-[60%] skew-x-[-48deg] skew-y-[14deg] scale-[0.675] p-4">
        <div :for={i <- 1..150} class="relative h-8 w-16 border-l border-slate-700">
          <div
            :for={j <- 1..100}
            id={"tile-#{i}-#{j}"}
            class={"tile relative h-8 w-16 border-r border-t border-slate-700 transition-colors duration-300 ease-out hover:duration-0 #{get_colors() |> Enum.random()}"}
          >
          </div>
        </div>
      </div>

      <h1 class="relative z-20 text-xl text-white md:text-4xl">
        Hi this is my website
      </h1>
    </div>
    """
  end

  def mount(_params, _session, socket) do
    Presence.track(self(), @channel_topic, socket.id, %{
      socket_id: socket.id,
      x: -1,
      y: -1
    })

    # TODO: use localstorage to keep state across page refreshes

    AppWeb.Endpoint.subscribe(@channel_topic)

    initial_users =
      Presence.list(@channel_topic) |> Enum.map(fn {_, data} -> data[:metas] |> List.first() end)

    {:ok,
     socket |> push_event("main-page-set-users", %{users: initial_users, socket_id: socket.id}),
     layout: false}
  end

  def handle_event("main-page-mousemove", %{"x" => x, "y" => y}, socket) do
    IO.inspect({x, y, socket.id})

    metas =
      Presence.get_by_key(@channel_topic, socket.id)[:metas]
      |> List.first()
      |> Map.merge(%{x: x, y: y})

    Presence.update(self(), @channel_topic, socket.id, metas)

    {:noreply, socket}
  end

  def handle_info(%{event: "presence_diff", payload: _payload}, socket) do
    users =
      Presence.list(@channel_topic) |> Enum.map(fn {_, data} -> data[:metas] |> List.last() end)

    IO.inspect(Presence.list(@channel_topic))

    {:noreply, socket |> push_event("main-page-set-users", %{users: users, socket_id: socket.id})}
  end
end
