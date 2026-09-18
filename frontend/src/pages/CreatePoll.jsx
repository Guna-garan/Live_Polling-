import { useState } from "react";
import { useNavigate } from "react-router-dom";
import Input from "../components/Input";
import Button from "../components/Button";
import ErrorMessage from "../components/ErrorMessage";
import { pollApi } from "../services/api";
import { expiryOptions, expiresAtFromMinutes } from "../utils/formatters";

export default function CreatePoll() {
  const navigate = useNavigate();
  const [question, setQuestion] = useState("");
  const [options, setOptions] = useState(["", ""]);
  const [expiryMinutes, setExpiryMinutes] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  function updateOption(i, value) {
    setOptions((opts) => opts.map((o, idx) => (idx === i ? value : o)));
  }

  function addOption() {
    if (options.length >= 10) return;
    setOptions((opts) => [...opts, ""]);
  }

  function removeOption(i) {
    if (options.length <= 2) return;
    setOptions((opts) => opts.filter((_, idx) => idx !== i));
  }

  async function handleSubmit(e) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const minutes = expiryMinutes ? Number(expiryMinutes) : null;
      const poll = await pollApi.create({
        question,
        options: options.filter((o) => o.trim() !== ""),
        expiresAt: expiresAtFromMinutes(minutes),
      });
      navigate(`/poll/${poll.id}`, { state: { justCreated: true } });
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="create-poll-page">
      <form className="create-poll-card" onSubmit={handleSubmit}>
        <h1>Create a Poll</h1>
        <ErrorMessage>{error}</ErrorMessage>

        <Input
          label="Question"
          value={question}
          onChange={(e) => setQuestion(e.target.value)}
          maxLength={200}
          placeholder="What do you want to ask?"
          required
        />

        <div className="field">
          <span className="field-label">Options</span>
          {options.map((opt, i) => (
            <div className="option-row" key={i}>
              <input
                className="input"
                value={opt}
                onChange={(e) => updateOption(i, e.target.value)}
                placeholder={`Option ${i + 1}`}
                maxLength={100}
                required
              />
              {options.length > 2 && (
                <button
                  type="button"
                  className="option-remove"
                  onClick={() => removeOption(i)}
                  aria-label={`Remove option ${i + 1}`}
                >
                  ✕
                </button>
              )}
            </div>
          ))}
          {options.length < 10 && (
            <Button type="button" className="add-option" onClick={addOption}>
              + Add option
            </Button>
          )}
        </div>

        <label className="field">
          <span className="field-label">Poll duration</span>
          <select
            className="input"
            value={expiryMinutes}
            onChange={(e) => setExpiryMinutes(e.target.value)}
          >
            {expiryOptions().map((opt) => (
              <option key={opt.label} value={opt.minutes ?? ""}>
                {opt.label}
              </option>
            ))}
          </select>
        </label>

        <Button type="submit" loading={loading} className="btn-block">
          Create Poll
        </Button>
      </form>
    </div>
  );
}
