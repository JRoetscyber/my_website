# SEO Canonical Domain Guide

## The Goal

Pick one official version of the website and make every SEO signal point to it.

For this site, the recommended canonical domain is:

```text
https://jo4.dev
```

That means Google, Bing, your sitemap, your canonical tags, and any alternate domains should all agree that `https://jo4.dev` is the main version.

## Why This Matters

Search engines do not always assume these are the same website:

```text
http://jo4.dev
https://jo4.dev
http://www.jo4.dev
https://www.jo4.dev
https://jo4.co.za
```

To a browser, they may all show the same site. To Google, they can look like separate URLs unless redirects and canonical tags are set correctly.

If the same page is available on multiple domains or protocols, Google may split SEO value between them. That can weaken rankings because backlinks, indexed pages, sitemap links, and user signals are not concentrated on one version.

Example problem:

```text
https://jo4.dev/services
https://jo4.co.za/services
http://jo4.dev/services
```

If all three load the same page without redirecting, Google has to decide which one is the real page. That is not ideal. You want to make the decision obvious.

## What Has Already Been Fixed In The App

These changes have already been made in the code:

1. The sitemap now generates HTTPS links.
2. `robots.txt` now points to:

```text
https://jo4.dev/sitemap.xml
```

3. Canonical tags now use HTTPS.
4. Structured-data URLs now use HTTPS.

These are good on-page SEO signals, but server/domain redirects still need to be checked outside the Flask code.

## What Still Needs To Be Done

### 1. Choose The Main Domain

Use this as the main domain:

```text
https://jo4.dev
```

Do not mix `jo4.dev` and `jo4.co.za` in SEO tools unless one redirects to the other.

### 2. Redirect HTTP To HTTPS

This should happen:

```text
http://jo4.dev
```

Redirects to:

```text
https://jo4.dev
```

This should be a `301` permanent redirect.

### 3. Redirect WWW To Non-WWW

If `www.jo4.dev` is active, it should redirect to the non-www version:

```text
https://www.jo4.dev
```

Redirects to:

```text
https://jo4.dev
```

This should also be a `301` permanent redirect.

### 4. Redirect Old Or Alternate Domains

If `jo4.co.za` is still connected to the same website, it should redirect to `jo4.dev`.

Example:

```text
https://jo4.co.za/services
```

Should redirect to:

```text
https://jo4.dev/services
```

This keeps the same path, but moves it to the main domain.

### 5. Submit Only The Main Sitemap

In Google Search Console, submit:

```text
https://jo4.dev/sitemap.xml
```

Do not submit a second sitemap for `jo4.co.za` unless that domain is meant to be a separate website.

### 6. Check Google Search Console Properties

The cleanest setup is a domain property for:

```text
jo4.dev
```

You can also add a URL-prefix property for:

```text
https://jo4.dev
```

The important part is that indexing and sitemap submission should focus on the canonical domain.

## How To Test Redirects

After deployment, check these URLs in a browser:

```text
http://jo4.dev
https://www.jo4.dev
http://www.jo4.dev
https://jo4.co.za
```

They should all end up at:

```text
https://jo4.dev
```

For inner pages, test this too:

```text
https://jo4.co.za/services
```

It should end up at:

```text
https://jo4.dev/services
```

## What A Correct Setup Looks Like

Good:

```text
https://jo4.dev/services
```

Canonical tag:

```html
<link rel="canonical" href="https://jo4.dev/services">
```

Sitemap:

```xml
<loc>https://jo4.dev/services</loc>
```

Bad:

```text
http://jo4.dev/services
https://www.jo4.dev/services
https://jo4.co.za/services
```

All loading the same content without redirecting.

## Simple Explanation

Think of the canonical domain as the official address of the business.

If customers, Google, and other websites use five different addresses for the same business, the reputation gets spread out. If everyone uses one address, the reputation builds in one place.

For SEO, that one address should be:

```text
https://jo4.dev
```

## Final Checklist

- [ ] Main domain chosen: `https://jo4.dev`
- [ ] Sitemap uses `https://jo4.dev`
- [ ] `robots.txt` points to `https://jo4.dev/sitemap.xml`
- [ ] Canonical tags use `https://jo4.dev`
- [ ] HTTP redirects to HTTPS
- [ ] `www` redirects to non-www
- [ ] `jo4.co.za` redirects to `jo4.dev`, if it points to the same site
- [ ] Google Search Console sitemap submitted for `https://jo4.dev/sitemap.xml`
- [ ] Google indexes only the `https://jo4.dev` versions of pages
