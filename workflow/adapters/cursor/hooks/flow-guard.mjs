// agent-flow: the agent may never approve or reject. The human runs these
// commands in their own terminal. "deny" is used because "ask" is reported
// to be ignored in some Cursor versions.
let input = "";
process.stdin.on("data", (chunk) => (input += chunk));
process.stdin.on("end", () => {
  const { command = "" } = JSON.parse(input || "{}");
  const humanOnly = /\bflow(\.exe)?\s+(approve|reject)\b/;
  if (humanOnly.test(command)) {
    console.log(JSON.stringify({
      permission: "deny",
      user_message: "Only you can run this. Type it in your own terminal: " + command,
      agent_message: "Only the human may approve or reject. Ask them to run the command themselves.",
    }));
  } else {
    console.log(JSON.stringify({ permission: "allow" }));
  }
});
