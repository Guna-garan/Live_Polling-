import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import Results from "./Results";

const options = [
  { id: "a", text: "React" },
  { id: "b", text: "Vue" },
];

describe("Results", () => {
  it("renders each option with its vote count and percentage", () => {
    render(<Results options={options} results={{ a: 3, b: 1 }} totalVotes={4} />);
    expect(screen.getByText("React")).toBeInTheDocument();
    expect(screen.getByText("75%")).toBeInTheDocument();
    expect(screen.getByText("25%")).toBeInTheDocument();
    expect(screen.getByText("4 total votes")).toBeInTheDocument();
  });

  it("calls onSelect with the option id when selectable and clicked", () => {
    const onSelect = vi.fn();
    render(
      <Results
        options={options}
        results={{}}
        totalVotes={0}
        selectable
        selectedOptionId={null}
        onSelect={onSelect}
      />
    );
    fireEvent.click(screen.getByText("Vue"));
    expect(onSelect).toHaveBeenCalledWith("b");
  });
});
