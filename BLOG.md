# From a Broken Camera to a Box of Old Phones

**How I turned a pile of retired Androids and a 2014 Mac mini into a CCTV system — and gave e-waste a second life.**

---

## The moment it all started

There's a corner in my warehouse that most people would call *junk*.

Old Android tablets. Retired phones. Scratched screens, tired batteries — every one of them long past its "upgrade cycle." They sat in a quiet pile: not broken enough to throw away, not useful enough to keep. Just… waiting.

That same week, the camera at home died.

I stood there, staring at that box of forgotten devices, and it hit me:

> *I don't have a broken camera problem. I have an inventory problem.*

Every single phone in that box has a working camera, a working Wi-Fi chip, and a tiny computer once powerful enough to run an entire phone in your pocket. And I was about to buy a brand-new CCTV camera — when a dozen cameras were already gathering dust in my own warehouse.

So I didn't buy one.

I built something instead.

---

## The idea

What if every retired Android device became a camera in a CCTV system I actually *own*?

- One small server runs on a single machine.
- An Android app is installed on each old phone.
- The phone's camera streams live video up to the server.
- From there, I watch every feed on my laptop — or on my personal phone — exactly like a real CCTV dashboard.

No cloud. No monthly fee. No company watching my home from a datacenter somewhere.

And the best part? These old phones stream with their **screens completely off**. A retired phone, sleeping on a shelf, quietly watching over the house.

---

## The build

The server is written in **Go** — a single, self-contained binary that runs on Windows, macOS, and Linux. No runtime, no dependencies, no Docker. Just one file.

Each old Android device runs **HuzHome**, the companion app. It registers with the server and streams live video over **WebRTC**, peer-to-peer. The server only relays signaling — it never touches the video itself. The browser connects straight to the camera phone.

The result is a dashboard where I can watch all my cameras at once, switch between views, zoom in, pan and tilt the frame, toggle motion detection, snap a photo — all from my couch, on my own phone.

The server doesn't just work. It barely even *notices* the work.

---

## The test that surprised me

Here's the part I didn't expect.

The whole system runs on a **2014 Mac mini** — the kind of machine most people would also call "retired." And the cameras? A handful of Android phones from a decade ago.

I streamed continuously from **two devices**, day and night, and then checked the server:

| Metric | Value | Verdict |
|---|---|---|
| CPU usage | ~4.6% | 🟢 Very low |
| CPU idle | 95.3% | 🟢 Excellent |
| Load (1 min) | 0.24 | 🟢 Very low |
| Load (5 min) | 0.05 | 🟢 Very low |
| Load (15 min) | 0.02 | 🟢 Very low |
| CPU cores | 4 | 🟢 More than enough |
| RAM | 644 MB / 3.2 GB | 🟢 Good |
| RAM available | 2.6 GB | 🟢 Excellent |
| Swap | ~0 MB | 🟢 No issues |
| Zombie processes | 0 | 🟢 Healthy |

**4.6% CPU while streaming two live cameras. On a 2014 Mac mini.**

Two cameras, a web dashboard, device discovery, authentication — the whole system idling at 95% free CPU, sitting comfortably inside 3.2 GB of RAM. Swap barely touched. Not a single zombie process.

The "weak" machine wasn't weak at all.

It was just waiting for a job worth doing.

---

## Why this matters to me

Every year, millions of phones end up in drawers, and then in landfills — perfectly capable hardware, retired not because it broke, but because we were told it was *old*.

This project is a small act of rebellion against that.

It's a 2014 Mac mini and a box of forgotten Androids, saying: **"We're still here. We're still useful. And we can still watch over your home."**

The camera at home got fixed — not by buying a new one, but by remembering that I already owned cameras. Lots of them. They were just waiting in a box, their screens dark, their cameras still awake.

---

## Try it yourself

The project is **free and open-source**:

- **Huz CCTV Server** — Go backend, single binary, cross-platform (Windows / macOS / Linux)
- **HuzHome** — the Android companion app that turns old phones into cameras

Watch your cameras from any browser, control them remotely, never pay a subscription. All you need is one box of old phones, one small server, and one good idea.
