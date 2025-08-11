- content in the output is eventually truncated ...[earlier messages truncated]... ... it shouldnt be.

- i would like to be able to push up in the input and iterate through previous messages sent.. eg, so i can push up arrow on my keyboard and the input populates with my last sent message, and if i hit up again i get the second last sent message etc. add history to the text input, so the user can hit up key to populate the input with their last prompt.

- while typing in the input, sometimes the text content visible in the history moves up and down a little bit. that should not happen.

- while an agent is working, we should display an elapsed time counter for that agent in the sidebar, perhaps we use something like this: https://github.com/charmbracelet/bubbletea/blob/main/examples/stopwatch/main.go

- when agentry loads, just below the current agentry logo, title and version number.. we should also display whether agent 0 prompt file is present and whether the user has at least one api key defined in env variables.

- i want to use the little minidots spinner when an agent is active, like this: https://github.com/charmbracelet/bubbletea/blob/main/examples/spinners/main.go

- i want to replace the rotating "/" slash style spinners at the end of ongoing messages to use the 3 dots spinner https://github.com/charmbracelet/bubbletea/blob/main/examples/spinners/main.go


- left/right should be reserved for moving the cursor around the text input, i think we need a different way to switch agents, maybe shift + left/right?

- markup code highlighting formatting

- nerdfonts glyphs? https://github.com/ryanoasis/nerd-fonts