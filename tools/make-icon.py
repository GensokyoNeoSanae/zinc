#!/usr/bin/env python3
"""生成 zinc.ico：透明底 + 斜体 Zn + 银白描边。

每个尺寸单独渲染：8 倍超采样画「字形 + 描边」，裁到墨迹框，再缩到目标尺寸。
不这么做的话，16px 是从 256px 缩下来的，字母会糊成一团。

只依赖 Pillow：
    python tools/make-icon.py            # 写到仓库根目录的 zinc.ico
    python tools/make-icon.py --out x.ico --preview out.png

设计约束（2026-10-27 定）：
- 只留符号 Zn，斜体（额外 6° 倾斜，强调"速度感"）
- 背景透明，深浅背景都能看清 → 钢灰填充 + 较粗的银白描边
- 全程纯色，不用任何渐变
"""
import argparse
import io
import math
import struct
import sys

from PIL import Image, ImageDraw, ImageFont

FONT = r"C:\Windows\Fonts\segoeuiz.ttf"  # Segoe UI Bold Italic

SIZES = [16, 20, 24, 32, 40, 48, 64, 96, 128, 256]
SS = 8  # 超采样倍数

FILL = (0xF1, 0xF5, 0xF9)      # 银白：深色背景上靠它跳出来
STROKE = (0x2F, 0x38, 0x43)    # 深灰描边：浅色背景上靠它勾边
STROKE_RATIO = 0.115           # 描边宽度 / 画布边长（≥48px）
SHEAR_DEG = 6.0                # 在斜体之上再加的角度
GLYPH_BOX = 0.76               # 字形占画布宽度的比例
MARGIN = 0.94                  # 裁切后放大到 94%，留一点呼吸空间
SMALL_STROKE_SCALE = 0.45      # ≤32px 的描边缩放：环太粗会把字吃掉


def _hex(text):
    t = text.lstrip("#")
    return tuple(int(t[i:i + 2], 16) for i in (0, 2, 4))


def glyph_layer(size):
    """把「Zn + 描边」画在 (size*SS) 的画布上，留足边距避免切边。"""
    S = size * SS
    pad = round(S * 0.30)
    canvas = Image.new("RGBA", (S + pad * 2, S + pad * 2), (0, 0, 0, 0))
    draw = ImageDraw.Draw(canvas)

    font_size = round(S * 0.90)
    while font_size > 8:
        font = ImageFont.truetype(FONT, font_size)
        box = draw.textbbox((0, 0), "Zn", font=font)
        if (box[2] - box[0]) <= S * GLYPH_BOX:
            break
        font_size -= max(1, font_size // 24)
    font = ImageFont.truetype(FONT, font_size)

    # 小尺寸按比例收描边：16px 上 1.5px 的环会把字吃掉，看起来是五个色块
    if size >= 48:
        scale = 1.0
    elif size >= 33:
        scale = 0.72
    else:
        scale = SMALL_STROKE_SCALE
    stroke_width = max(1, round(S * STROKE_RATIO * scale))
    draw.text((pad, pad), "Zn", font=font, fill=FILL + (255,),
              stroke_width=stroke_width, stroke_fill=STROKE + (255,))

    if SHEAR_DEG:
        k = math.tan(math.radians(SHEAR_DEG))
        w, h = canvas.size
        canvas = canvas.transform(
            (w + round(abs(k) * h), h), Image.AFFINE,
            (1, -k, k * h, 0, 1, 0), resample=Image.BICUBIC)
    return canvas


def render(size):
    layer = glyph_layer(size)
    box = layer.getbbox()
    if not box:
        return Image.new("RGBA", (size, size), (0, 0, 0, 0))
    glyph = layer.crop(box)
    w, h = glyph.size
    scale = min(size * MARGIN / w, size * MARGIN / h)
    nw, nh = max(1, round(w * scale)), max(1, round(h * scale))
    glyph = glyph.resize((nw, nh), Image.LANCZOS)
    out = Image.new("RGBA", (size, size), (0, 0, 0, 0))
    out.paste(glyph, ((size - nw) // 2, (size - nh) // 2), glyph)
    return out


def write_ico(path, frames):
    """自己拼 ICO：每帧一张 PNG（Vista 以后都支持），保留每尺寸的精修结果。"""
    blobs = []
    for frame in frames:
        buf = io.BytesIO()
        frame.save(buf, format="PNG", optimize=True)
        blobs.append(buf.getvalue())

    header = struct.pack("<HHH", 0, 1, len(frames))
    offset = len(header) + 16 * len(frames)
    entries = b""
    for frame, blob in zip(frames, blobs):
        w, h = frame.size
        entries += struct.pack(
            "<BBBBHHII",
            0 if w >= 256 else w, 0 if h >= 256 else h,
            0, 0, 1, 32, len(blob), offset)
        offset += len(blob)

    with open(path, "wb") as f:
        f.write(header + entries + b"".join(blobs))


def main():
    global FILL, STROKE, STROKE_RATIO, SHEAR_DEG

    ap = argparse.ArgumentParser()
    ap.add_argument("--out", default="zinc.ico")
    ap.add_argument("--preview", default="")
    ap.add_argument("--fill", default="#F1F5F9", help="字形颜色，如 #F1F5F9")
    ap.add_argument("--stroke", default="#2F3843", help="描边颜色")
    ap.add_argument("--stroke-ratio", type=float, default=0.095)
    ap.add_argument("--shear", type=float, default=6.0)
    args = ap.parse_args()

    FILL = _hex(args.fill)
    STROKE = _hex(args.stroke)
    STROKE_RATIO = args.stroke_ratio
    SHEAR_DEG = args.shear

    frames = [render(s) for s in SIZES]
    write_ico(args.out, frames)

    if args.preview:
        # 预览：256 大图 + 实际尺寸一行，直接看各尺寸是不是还能认出 Zn
        pad, gap = 12, 10
        strip = [256] + SIZES
        width = 256 + pad * 2 + 24 + sum(s for s in SIZES) + gap * (len(SIZES) - 1)
        height = 256 + pad * 2
        canvas = Image.new("RGB", (width, height), (246, 247, 249))
        canvas.paste(render(256), (pad, pad), render(256))
        x = pad + 256 + 24
        for s in SIZES:
            img = render(s)
            canvas.paste(img, (x, pad + 256 - s), img)
            x += s + gap
        canvas.save(args.preview)

    print(f"wrote {args.out} ({len(SIZES)} frames: {', '.join(map(str, SIZES))})")


if __name__ == "__main__":
    sys.exit(main())
