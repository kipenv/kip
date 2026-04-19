from PIL import Image, ImageDraw, ImageFilter, ImageChops
import math, os, subprocess

W, H   = 100, 660
FPS    = 20
TOTAL  = 26.0

ARROW_SIZE  = 8
ARROW_XS    = [20, 50, 80]
ARROW_Y     = H // 2
CYCLE       = 2.5

# push typing ends at ~6.3s → arrows appear as it processes
APPEAR_T    = 6.3
APPEAR_DUR  = 0.7
FADE_T      = 13.6   # receiver starts typing (Sleep 13600ms)
FADE_DUR    = 0.5

BG     = (15,  23,  42)
DIM    = (35,  50,  90)
BRIGHT = (122, 162, 247)

def smoothstep(t):
    t = max(0.0, min(1.0, t))
    return t * t * (3 - 2 * t)

def clamp01(t, start, dur):
    return smoothstep((t - start) / dur)

def lerp(a, b, t):
    return tuple(int(a[i] + (b[i] - a[i]) * t) for i in range(3))

def draw_chevron(draw, cx, cy, size, color, alpha):
    half = max(1, size // 2)
    pts = [(cx - half, cy - size), (cx + half, cy), (cx - half, cy + size)]
    draw.polygon(pts, fill=(*color, min(255, max(0, alpha))))

FEATHER = 90

def make_reveal_mask(reveal_frac, overall_alpha):
    mask = Image.new('L', (W, H), 0)
    if overall_alpha <= 0 or reveal_frac <= 0:
        return mask
    frontier_y = int(H * (1.0 - reveal_frac))
    pixels = mask.load()
    for y in range(H):
        if y < frontier_y:
            val = 0
        else:
            dist = y - frontier_y
            edge_t = min(1.0, dist / FEATHER)
            val = int(255 * smoothstep(edge_t) * overall_alpha)
        pixels[0, y] = val
    for x in range(1, W):
        for y in range(H):
            pixels[x, y] = pixels[0, y]
    return mask

FRAMES = int(FPS * TOTAL) + 1
OUT_DIR = '/tmp/arrow_frames'
os.makedirs(OUT_DIR, exist_ok=True)

for i in range(FRAMES):
    t = i / FPS
    appear  = clamp01(t, APPEAR_T, APPEAR_DUR)
    fade    = 1.0 - clamp01(t, FADE_T, FADE_DUR)
    overall = appear * fade

    # Transparent background — ffmpeg overlay handles compositing
    frame = Image.new('RGBA', (W, H), (0, 0, 0, 0))

    if overall > 0.01:
        reveal_frac = clamp01(t, APPEAR_T, APPEAR_DUR)
        arrow_layer = Image.new('RGBA', (W, H), (0, 0, 0, 0))

        for j, ax in enumerate(ARROW_XS):
            phase = (t / CYCLE - j / len(ARROW_XS)) % 1.0
            b = ((math.sin(phase * 2 * math.pi - math.pi / 2) + 1) / 2) ** 0.55

            glow = Image.new('RGBA', (W, H), (0, 0, 0, 0))
            draw_chevron(ImageDraw.Draw(glow), ax, ARROW_Y, ARROW_SIZE + 3, BRIGHT, int(140 * b))
            arrow_layer = Image.alpha_composite(arrow_layer, glow.filter(ImageFilter.GaussianBlur(6)))

            sharp = Image.new('RGBA', (W, H), (0, 0, 0, 0))
            draw_chevron(ImageDraw.Draw(sharp), ax, ARROW_Y, ARROW_SIZE, lerp(DIM, BRIGHT, b), int(60 + 195 * b))
            arrow_layer = Image.alpha_composite(arrow_layer, sharp)

        mask = make_reveal_mask(reveal_frac, overall)
        arrow_layer.putalpha(ImageChops.multiply(arrow_layer.getchannel('A'), mask))
        frame = Image.alpha_composite(frame, arrow_layer)

    frame.save(f'{OUT_DIR}/frame_{i:04d}.png')
    if i % 50 == 0:
        print(f"  frame {i}/{FRAMES}  t={t:.1f}s")

# Convert PNG sequence → video with alpha channel
subprocess.run([
    'ffmpeg', '-y',
    '-framerate', str(FPS),
    '-i', f'{OUT_DIR}/frame_%04d.png',
    '-c:v', 'qtrle', '-pix_fmt', 'argb',
    'assets/arrows.mov'
], check=True)
print(f"Done → assets/arrows.mov  ({FRAMES} frames)")
