# Bug reproduction

Recipe policies and provider returns share mutable maps and required slices. A typed-nil provider also passes an interface nil check, allowing invalid policy access.

Run the targeted recipe test. It shows that modifying a cloned policy changes the original cached policy.
