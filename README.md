# Untouch

Untouch is a drop-in replacement for the touchscreen interface of
certain digital and hybrid piano models made by Kawai.

The Untouch user interface comprises an 8-letter 14-segment display
and nine input buttons. Occasionally, the 88-key piano keyboard is
being used for additional input.

It has been devoloped on a Kawai NV10 hybrid piano, but can be
expected to work on the CA78 and CA98 digital pianos as well.

## Design decisions

- Boot up faster than the piano mainboard.

- On startup, always load registration 0. Never be outside a
  registration.

- No input menus, only short key sequences.

- Switch off the display whenever possible.
