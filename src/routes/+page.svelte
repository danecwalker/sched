<script lang="ts">
  import type { PageProps } from "./$types";
  import { Gradient } from "$lib/gradient.js";
  import { onMount } from "svelte";
  import { Clock, Donut, Loader2 } from "lucide-svelte";
  import { getShifts, pb, type ShiftCollection } from "$lib/pb";

  let now = $state(new Date(Date.now() + 24*60*60*1000));
  let loaded = $state(false);
  let myshifts: { current: ShiftCollection[]; next: ShiftCollection[] } =
    $state({ current: [], next: [] });
  let currentView: "current" | "next" = $state("current");
  let total_hours = $derived.by(() => {
    let total = 0;
    myshifts[currentView].forEach((shift) => {
      total += shift.duration;
    });
    return total;
  });
  let total_unpaid_hours = $derived.by(() => {
    let total = 0;
    myshifts[currentView].forEach((shift) => {
      total += shift.unpaid_break_duration;
    });
    return total;
  });

  let displayPay = $state(false);

  let DateFormatter = new Intl.DateTimeFormat('en', {
    timeZone: 'Australia/Brisbane',
    weekday: 'long',
    month: 'long',
    day: 'numeric',
  });

  onMount(() => {
    const gradient = new Gradient();
    gradient.initGradient("#canvas");

    pb.collection("shifts").subscribe("*", async (data) => {
      myshifts = await getShifts();
    });

    if (!loaded)
      getShifts().then((shifts) => {
        myshifts = shifts;
        loaded = true;
      });

    let interval = setInterval(() => {
      now = new Date(Date.now() + 4*24*60*60*1000 - 13*60*60*1000 - 42*60*1000);
      // find today
      const today = myshifts.current.find((shift) => {
        return shift.date === DateFormatter.format(now);
      });

      if (today) {
        const current = today.shifts.find((block) => {
          // now is within 5 minutes of the start of the block
          let start = new Date(block.raw_start_time);
          return (
            now >= new Date(start.getTime() - 5*60*1000) &&
            now <= start
          );
        });

        if (current) {
          if (localStorage.getItem(current.id) !== "true") {
            new Notification("Shift Reminder", {
              body: `Your ${capitalize(current.name)} shift begins soon.`,
              icon: "/icon-512.png"
            });
            localStorage.setItem(current.id, "true");
            console.log("Current block");
            console.log(current);
          }
        }
      }
    }, 1000);

    Notification.requestPermission().then((result) => {
      if (result === "granted") {
        console.log("Notification permission granted");
      }
    });

    return () => {
      clearInterval(interval);
    };
  });

  const capitalize = (str: string) => {
    let words = str.toLowerCase().split(" ");
    for (let i = 0; i < words.length; i++) {
      words[i] = words[i][0].toUpperCase() + words[i].slice(1);
    }
    return words.join(" ");
  };

  const prettyDuration = (
    duration: number,
    opt?: {
      full?: boolean;
    }
  ) => {
    if (opt?.full) {
      const hours = Math.floor(duration);
      const minutes = (duration * 60) % 60;
      let builder = "";
      if (hours > 0) {
        builder += `${hours}H `;
      }
      if (minutes > 0) {
        builder += `${minutes}M`;
      }
      if (builder === "") {
        builder = "";
      }
      return builder;
    } else {
      if (duration >= 1) {
        return `${duration}H`;
      }
      return `${duration * 60}M`;
    }
  };
</script>

<svelte:head>
  <link rel="manifest" href="site.webmanifest" />
  <title>My Roster</title>
  <meta name="theme-color" content="#f5f5f5" />
</svelte:head>

<div class="bg-neutral-100 w-full h-dvh">
  <div
    class="relative w-full h-full max-w-lg mx-auto flex flex-col pb-8 overflow-y-auto"
  >
    <div class="p-4 flex flex-col gap-4">
      <div
        class="relative w-full bg-neutral-100 aspect-[7/3] overflow-hidden rounded-md"
      >
        <canvas id="canvas" class="w-full h-full"></canvas>
        <div
          class="absolute top-0 left-0 w-full h-full flex flex-col justify-center items-center text-neutral-800"
        >
          {#if displayPay}
            <button
              onclick={() => (displayPay = !displayPay)}
              class="flex flex-col select-none justify-center items-center gap-2"
            >
              <div class="relative">
                <span class="text-6xl font-bold"
                  >{(total_hours - total_unpaid_hours) * 26}</span
                >
                <span
                  class="absolute text-2xl font-bold top-1 -left-1 -translate-x-full"
                  >$</span
                >
              </div>
              <div class="text-sm flex gap-3 text-neutral-600">
                <div class="flex items-center gap-1">
                  <Clock size={16} />
                  <span>{prettyDuration(total_hours - total_unpaid_hours)}</span
                  >
                </div>
                {#if total_unpaid_hours > 0}
                  <div class="flex items-center gap-1">
                    <Donut size={16} />
                    <span>{prettyDuration(total_unpaid_hours)}</span>
                  </div>
                {/if}
              </div>
            </button>
          {:else}
            <button
              onclick={() => (displayPay = !displayPay)}
              class="flex flex-col select-none justify-center items-center"
            >
              <span class="text-6xl font-bold"
                >{prettyDuration(total_hours)}</span
              >
            </button>
          {/if}
        </div>
      </div>

      <div class="w-full flex justify-start items-center gap-2">
        <button onclick={() => (currentView = "current")}
          ><div
            class={`text-xs 
        ${currentView === "current" ? "text-neutral-100 bg-neutral-950 border-neutral-950" : "text-neutral-500 bg-neutral-200 border-neutral-300"} 
        py-2 px-4 rounded-full cursor-pointer border transition duration-200`}
          >
            Current Week
          </div></button
        >
        <button onclick={() => (currentView = "next")}
          disabled={myshifts.next.length === 0}
          ><div
            class={`text-xs 
          ${currentView !== "current" ? "text-neutral-100 bg-neutral-950 border-neutral-950" : "text-neutral-500 bg-neutral-200 border-neutral-300"} 
          py-2 px-4 rounded-full cursor-pointer border transition duration-200`}
          >
            {
              myshifts.next.length === 0
                ? "Unavailable"
                : `Next Week`
            }
          </div></button
        >
      </div>

      {#if !loaded}
        <div class="w-full flex justify-center items-center text-neutral-400">
          <Loader2 class="animate-spin" />
        </div>
      {:else if myshifts[currentView].length === 0}
        <div class="w-full flex justify-center items-center text-neutral-400">
          <span>No shifts found</span>
        </div>
      {:else}
        {#each myshifts[currentView] as shift}
          <div
            class="w-full bg-neutral-50 overflow-hidden rounded-md border border-neutral-400 text-neutral-500"
          >
            <div
              class="w-full bg-neutral-200 p-2 border-b border-neutral-400 text-sm flex justify-between"
            >
              <span>{shift.date}</span>
              <div class="text-sm flex gap-3">
                <div class="flex items-center gap-1">
                  <Clock size={16} />
                  <span>{prettyDuration(shift.duration, { full: true })}</span>
                </div>
                {#if shift.break_duration > 0}
                  <div class="flex items-center gap-1">
                    <Donut size={16} />
                    <span
                      >{prettyDuration(shift.break_duration, {
                        full: true,
                      })}</span
                    >
                  </div>
                {/if}
              </div>
            </div>
            <div class="flex flex-col gap-2 p-2">
              {#each shift.shifts as block}
                <div
                  class={`w-full p-2 flex justify-between rounded-md border transition duration-200 ${
                    now >= new Date(block.raw_end_time)
                      ? "border-neutral-300"
                      : block.name.toUpperCase() === "FRESH PRODUCE"
                        ? "border-emerald-600"
                        : block.name.toUpperCase() === "DAIRY & FROZEN"
                          ? "border-sky-600"
                          : block.name.toUpperCase() === "OVERHEAD"
                            ? "border-violet-600"
                            : block.name.toUpperCase() === "MEAT"
                              ? "border-red-600"
                              : "border-neutral-600"
                  } ${
                    now >= new Date(block.raw_end_time)
                      ? "bg-neutral-100"
                      : block.name.toUpperCase() === "FRESH PRODUCE"
                        ? "bg-emerald-200"
                        : block.name.toUpperCase() === "DAIRY & FROZEN"
                          ? "bg-sky-200"
                          : block.name.toUpperCase() === "OVERHEAD"
                            ? "bg-violet-200"
                            : block.name.toUpperCase() === "MEAT"
                              ? "bg-red-200"
                              : "bg-neutral-200"
                  } ${
                    now >= new Date(block.raw_end_time)
                      ? "text-neutral-300"
                      : block.name.toUpperCase() === "FRESH PRODUCE"
                        ? "text-emerald-800"
                        : block.name.toUpperCase() === "DAIRY & FROZEN"
                          ? "text-sky-800"
                          : block.name.toUpperCase() === "OVERHEAD"
                            ? "text-violet-800"
                            : block.name.toUpperCase() === "MEAT"
                              ? "text-red-800"
                              : "text-neutral-800"
                  }`}
                >
                  <div class="flex flex-col">
                    <span>{block.name}</span>
                    <div class="text-sm flex gap-3">
                      <div class="flex items-center gap-1">
                        <Clock size={16} />
                        <span
                          >{prettyDuration(block.duration, {
                            full: true,
                          })}</span
                        >
                      </div>
                    </div>
                  </div>
                  <div class="flex flex-col">
                    <span class="font-bold">{block.start_time}</span>
                    <span
                      class={now >= new Date(block.raw_end_time)
                        ? "text-neutral-300"
                        : block.name.toUpperCase() === "FRESH PRODUCE"
                          ? "text-emerald-600"
                          : block.name.toUpperCase() === "DAIRY & FROZEN"
                            ? "text-sky-600"
                            : block.name.toUpperCase() === "OVERHEAD"
                              ? "text-violet-600"
                              : block.name.toUpperCase() === "MEAT"
                                ? "text-red-600"
                                : "border-neutral-600"}>{block.end_time}</span
                    >
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/each}
      {/if}
    </div>

    <!-- {#if updates}
      <div class="absolute bottom-0 left-0 w-full flex justify-center items-center p-4">
        <button onclick={refreshShifts}><div class="text-sm text-neutral-100 bg-neutral-950 py-2 px-4 rounded-full cursor-pointer">Updates Available</div></button>
      </div>
    {/if} -->
  </div>
</div>
